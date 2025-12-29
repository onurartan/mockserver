document.addEventListener("alpine:init", () => {
  Alpine.data("app", () => ({
    // Core state
    config: {},
    user: null,
    oas: {}, // OpenAPI Specification
    activeTab: "dashboard",

    // Route management
    routes: [],
    tags: [],
    search: "",

    // System status flags
    initialLoading: true,
    loading: true,
    globalError: null,
    logsLoading: false,

    // Auth & Logging
    authModalOpen: false,
    globalAuthStore: {},
    health: null,
    logs: [],

    toasts: [],

    /**
     * Initializes the application state.
     * Loads system configuration, handles deep linking via hash, and manages the initial loader.
     */
    async initApp() {
      this.initialLoading = true;
      this.globalError = null;

      try {
        await this.loadSystemState();

        // Handle deep linking (e.g., refreshing page on specific route)
        this.handleHashNavigation();
        window.addEventListener("hashchange", () =>
          this.handleHashNavigation()
        );

        setTimeout(() => (this.initialLoading = false), 500);
      } catch (e) {
        console.error("Init Error:", e);
        this.initialLoading = false;
        this.globalError =
          "Failed to connect to MockServer. Ensure the backend is running.";
      }
    },

    /**
     * Displays a temporary toast notification (Success/Error/Warning).
     * Uses a unique ID to manage the lifecycle (enter/leave animations) of each toast.
     */
    toast(message, type = "success") {
      const id = Date.now();
      const newToast = {
        id: id,
        message: message,
        type: type,
        show: false,
        time: new Date().toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
        }),
      };

      this.toasts.push(newToast);

      // Trigger enter animation
      this.$nextTick(() => {
        const index = this.toasts.findIndex((t) => t.id === id);
        if (index !== -1) this.toasts[index].show = true;
      });

      // Auto-dismiss after 4 seconds
      setTimeout(() => this.removeToast(id), 4000);
    },

    removeToast(id) {
      const index = this.toasts.findIndex((t) => t.id === id);
      if (index !== -1) {
        this.toasts[index].show = false;
        // Wait for leave animation to finish before removing from DOM
        setTimeout(() => {
          this.toasts = this.toasts.filter((t) => t.id !== id);
        }, 3000);
      }
    },

    async refreshData() {
      this.logsLoading = true;
      try {
        await this.loadSystemState();
        this.toast("System reloaded successfully");
      } catch (e) {
        console.error("Reload Error:", e);
        this.toast("Failed to reload system", "error");
      } finally {
        this.logsLoading = false;
      }
    },

    /**
     * The core data loader.
     * Fetches Configuration, OpenAPI spec, and User session in parallel for performance.
     * Also handles session expiration redirect logic.
     */
    async loadSystemState() {
      try {
        // Cache-busting is essential here to get the latest config state
        const resConfig = await fetch(
          "/console/mockserver.json?t=" + Date.now()
        );
        const contentType = resConfig.headers.get("content-type");

        // Detect session expiration (HTML response instead of JSON)
        if (contentType && contentType.includes("text/html")) {
          window.location.href = "/console/login?expired=true";
          return;
        }
        if (!resConfig.ok) throw new Error(`Server Error: ${resConfig.status}`);

        const cfg = await resConfig.json();

        // Parallel fetch for non-blocking UI load
        const [resOas, resMe] = await Promise.allSettled([
          fetch("/openapi.json?t=" + Date.now()),
          fetch("/console/me?t=" + Date.now()),
        ]);

        let oas = {};
        if (resOas.status === "fulfilled" && resOas.value.ok) {
          oas = await resOas.value.json();
        }

        if (resMe.status === "fulfilled" && resMe.value.ok) {
          const userData = await resMe.value.json();
          this.user = userData.user;
        }

        // Update reactive state
        this.config = cfg;
        this.oas = oas;
        this.syncAuthStore(cfg, oas);

        // Clear existing routes to force a clean re-render
        this.routes = [];
        this.tags = [];

        // Process routes and generate unique IDs for Alpine iteration keys
        await this.$nextTick();
        const newRoutes = (cfg.routes || []).map((r, i) => {
          const routeObj = { ...r };
          routeObj.uniqueId = `${i}_${Date.now()}`;

          if (!routeObj.tag || routeObj.tag.trim() === "") {
            routeObj.tag = "Default";
          }
          return routeObj;
        });

        // Organize tags alphabetically, keeping 'Default' first
        const newTags = [...new Set(newRoutes.map((r) => r.tag))].sort(
          (a, b) => {
            if (a === "Default") return -1;
            if (b === "Default") return 1;
            return a.localeCompare(b);
          }
        );

        this.routes = newRoutes;
        this.tags = newTags;

        // Fetch live monitoring data
        await Promise.all([this.fetchHealth(), this.fetchLogs()]);
      } catch (e) {
        this.toast(`Load System State Error: ${e}`, "error");
      } finally {
        this.loading = false;
      }
    },

    /**
     * Synchronizes global authentication schemes (API Keys, Bearer tokens).
     * Merges persisted local storage data with the current server configuration.
     */
    syncAuthStore(cfg, oas) {
      const savedAuth = localStorage.getItem("mockserver_global_auth");
      const authMap = savedAuth ? JSON.parse(savedAuth) : {};
      const newAuthStore = {};

      // Import global auth from MockServer config
      if (cfg.server.auth?.enabled) {
        const name = cfg.server.auth.name || "Authorization";
        const type = (cfg.server.auth.type || "apikey").toLowerCase();

        newAuthStore[name] = {
          value: authMap[name] || "",
          in: cfg.server.auth.in || "header",
          type: type,
          source: "config",
        };
      } else if (oas.components?.securitySchemes) {
        // Import schemes defined in OpenAPI
        Object.entries(oas.components.securitySchemes).forEach(
          ([key, scheme]) => {
            let name = scheme.name;
            let loc = scheme.in || "header";
            let type = (scheme.type || "").toLowerCase();

            if (type === "http" && scheme.scheme === "bearer") {
              name = "Authorization";
              loc = "header";
              type = "bearer";
            }

            if (name && !newAuthStore[name]) {
              newAuthStore[name] = {
                value: authMap[name] || "",
                in: loc,
                type: type,
                source: "openapi",
              };
            }
          }
        );
      }

      this.globalAuthStore = newAuthStore;
    },
    // syncAuthStore(cfg, oas) {
    //   const savedAuth = localStorage.getItem("mockserver_global_auth");
    //   const authMap = savedAuth ? JSON.parse(savedAuth) : {};
    //   const newAuthStore = {};

    //   // 1. Import schemes defined in OpenAPI
    //   if (oas.components?.securitySchemes) {
    //     Object.entries(oas.components.securitySchemes).forEach(
    //       ([key, scheme]) => {
    //         newAuthStore[scheme.name] = {
    //           value: authMap[scheme.name] || "",
    //           in: scheme.in,
    //           type: scheme.type,
    //         };
    //       }
    //     );
    //   }

    //   // 2. Import global auth from MockServer config
    //   if (cfg.server.auth?.enabled) {
    //     const k = cfg.server.auth.name || "Authorization";
    //     if (!newAuthStore[k]) {
    //       newAuthStore[k] = {
    //         value: authMap[k] || "",
    //         in: cfg.server.auth.in || "header",
    //         type: cfg.server?.auth?.type
    //       };
    //     }
    //   }

    //   console.log(newAuthStore)
    //   this.globalAuthStore = newAuthStore;
    // },
    saveAuth() {
      const dataToSave = {};
      Object.entries(this.globalAuthStore).forEach(([k, v]) => {
        dataToSave[k] = v.value;
      });
      localStorage.setItem(
        "mockserver_global_auth",
        JSON.stringify(dataToSave)
      );
      this.authModalOpen = false;
      this.toast("Global security settings saved");
    },

    clearAuth() {
      Object.keys(this.globalAuthStore).forEach(
        (k) => (this.globalAuthStore[k].value = "")
      );
      localStorage.removeItem("mockserver_global_auth");
      this.toast("Global security settings saved");
    },

    logout() {
      this.toast("Logging out securely...", "warning");
      setTimeout(() => (window.location.href = "/console/logout"), 200);
    },

    // Filter routes based on active tab (Tag) and search query
    get filteredRoutes() {
      let r = this.routes;
      if (this.activeTab !== "dashboard" && this.activeTab !== "logs")
        r = r.filter((x) => x.tag === this.activeTab);

      if (this.search && this.activeTab !== "logs") {
        const q = this.search.toLowerCase();
        r = r.filter(
          (x) =>
            x.path.toLowerCase().includes(q) ||
            x.name?.toLowerCase().includes(q)
        );
      }
      return r;
    },

    // async switchTab(tab) {
    //   this.activeTab = tab;
    //   if (tab === "logs") await this.fetchLogs();
    // },

    async switchTab(tab) {
      this.activeTab = tab;
      if (tab === "dashboard") window.location.hash = "dashboard";
      else if (tab === "logs") {
        window.location.hash = "logs";
        this.fetchLogs();
      } else {
        window.location.hash = `collections/${tab}`;
      }
    },

    // --- UI Formatting Helpers ---

    badgeLabel(t) {
      return {
        FETCH: "Proxy",
        MOCK_FILE: "Mock File",
        MOCK: "Mock",
        AUTH_GLOBAL: "Global Auth",
        AUTH_ROUTE: "Scoped Auth",
      }[t];
    },

    badgeColor(t) {
      return {
        FETCH: "bg-emerald-50 text-emerald-700 border-emerald-200",
        MOCK_FILE: "bg-blue-50 text-blue-700 border-blue-200",
        MOCK: "bg-indigo-50 text-indigo-700 border-indigo-200",
        AUTH_GLOBAL: "bg-rose-50 text-rose-700 border-rose-200",
        AUTH_ROUTE: "bg-purple-50 text-purple-700 border-purple-200",
      }[t];
    },

    async fetchHealth() {
      if (!this.config.server.debug?.enabled) return;
      try {
        const port = this.config.server.port;
        const debugPath = this.config.server.debug.path;
        const res = await fetch(`http://localhost:${port}${debugPath}/health`);
        this.health = await res.json();
      } catch (e) {
        console.error("Health fetch error:", e);
      }
    },

    async fetchLogs() {
      if (!this.config.server.debug?.enabled) return;

      try {
        const port = this.config.server.port;
        const debugPath = this.config.server.debug.path;
        const res = await fetch(
          `http://localhost:${port}${debugPath}/requests`
        );
        let data = await res.json();
        this.logs = data.map((l) => ({ ...l, expanded: false })).reverse();
      } catch (e) {
        console.error("Logs Fetch Failed:", e);
      }
    },

    hasGlobalAuth() {
      return Object.values(this.globalAuthStore).some(
        (x) => x.value.length > 0
      );
    },

    formatDate(str) {
      return str ? new Date(str).toLocaleString() : "-";
    },

    getTabTitle() {
      return this.activeTab === "dashboard"
        ? "Overview"
        : this.activeTab === "logs"
        ? "Traffic Logs"
        : this.activeTab;
    },

    formatUptime(str) {
      return str ? str.replace("s", "s").split(".")[0] + "s" : "-";
    },

    formatTime(str) {
      return str ? new Date(str).toLocaleTimeString() : "";
    },

    statusColor(s) {
      if (s >= 500) return { dot: "bg-rose-500", text: "text-rose-600" };
      if (s >= 400) return { dot: "bg-amber-500", text: "text-amber-600" };
      return { dot: "bg-emerald-500", text: "text-emerald-600" };
    },

    methodColor(m) {
      const c = {
        GET: "bg-blue-100 text-blue-700 border-blue-200",
        POST: "bg-emerald-100 text-emerald-700 border-emerald-200",
        PUT: "bg-amber-100 text-amber-700 border-amber-200",
        DELETE: "bg-rose-100 text-rose-700 border-rose-200",
      };
      return c[m] || "bg-slate-100 border-slate-200";
    },

    /**
     * Parses the URL hash to restore the UI state (Active Tab & Expanded Route).
     * Supported Format: #collections/{Tag}/route/{Method}/{Path}
     */
    handleHashNavigation() {
      const hash = window.location.hash;
      if (!hash || hash === "#dashboard") {
        this.activeTab = "dashboard";
        return;
      }

      if (hash === "#logs") {
        this.switchTab("logs");
        return;
      }

      // Format: #collections/Users/route/GET/users/stateful/{id}
      if (hash.startsWith("#collections/")) {
        const parts = hash.split("/route/");
        const tagPart = parts[0].replace("#collections/", "");
        this.activeTab = tagPart;

        if (parts[1]) {
          const subParts = parts[1].split("/");
          const method = subParts[0];
          const routePath = "/" + subParts.slice(1).join("/");

          // Dispatch event to open the specific route card
          this.$nextTick(() => {
            window.dispatchEvent(
              new CustomEvent("open-route", {
                detail: { path: routePath, method: method },
              })
            );
          });
        }
      }
    },
    // Sekme değiştirildiğinde Hash'i güncelle
    switchTab(tab) {
      this.activeTab = tab;
      if (tab === "dashboard") window.location.hash = "dashboard";
      else if (tab === "logs") {
        window.location.hash = "logs";
        this.fetchLogs();
      } else {
        window.location.hash = `collections/${tab}`;
      }
    },
  }));

  // ROUTE CARD COMPONENT
  // Handles individual route logic: params building, execution, and display
  Alpine.data("routeCard", (route, config, oas, globalAuthStore) => ({
    expanded: false,
    reqLoading: false,
    result: null,
    params: [],
    body: "",
    errors: {},
    errorMessage: "",
    badges: [],

    // Documentation & Execution details
    docs: { responses: [] },
    curlCmd: "",
    highlightedCurl: "",
    reqUrl: "",
    resHeaders: {},
    highlightedBody: "",
    activeCopy: null,

    init() {
      this.computeBadges();
      this.parseDocs();

      // Listen for global open requests (from hash navigation)
      window.addEventListener("open-route", (e) => {
        if (e.detail.path === route.path && e.detail.method === route.method) {
          if (!this.expanded) this.toggle();
          this.$el.scrollIntoView({
            behavior: "smooth",
            block: "center",
          });
        }
      });
    },

    /**
     * Extracts response documentation from the loaded OpenAPI spec.
     */
    parseDocs() {
      const fullPath = config.server.api_prefix + route.path;
      let oasDef = null;
      if (oas.paths) {
        let match = oas.paths[fullPath];
        // Normalize path params (/users/{id} vs /users/:id)
        if (!match) {
          const normalized = fullPath.replace(/:([a-zA-Z0-9_]+)/g, "{$1}");
          match = oas.paths[normalized];
        }
        if (match) oasDef = match[route.method.toLowerCase()];
      }

      if (oasDef && oasDef.responses) {
        this.docs.responses = Object.entries(oasDef.responses).map(
          ([code, def]) => ({
            code,
            description: def.description,
          })
        );
      }
    },

    /**
     * Generates a ready-to-use cURL command for the current request setup.
     * Includes headers, method, and body payload.
     */
    generateCurl(url, headers, method, body) {
      let cmd = `curl -X ${method} '${url}'`;
      Object.entries(headers).forEach(([k, v]) => {
        cmd += ` \\\n  -H '${k}: ${v}'`;
      });
      if (["POST", "PUT", "PATCH"].includes(method) && body) {
        const safeBody = body.replace(/'/g, "'\\''"); // Escape single quotes
        cmd += ` \\\n  -d '${safeBody}'`;
      }
      this.curlCmd = cmd;

      this.highlightedCurl = Prism.highlight(cmd, Prism.languages.bash, "bash");
    },

    toggle() {
      this.expanded = !this.expanded;
      if (this.expanded) {
        this.buildParams();

        // Update URL hash for shareable link
        const cleanPath = route.path.startsWith("/")
          ? route.path.substring(1)
          : route.path;
        window.location.hash = `collections/${route.tag}/route/${route.method}/${cleanPath}`;
      } else {
        window.location.hash = `collections/${route.tag}`;
      }
    },

    methodColor(m) {
      const c = {
        GET: "bg-blue-100 text-blue-700 border-blue-200",
        POST: "bg-emerald-100 text-emerald-700 border-emerald-200",
        PUT: "bg-amber-100 text-amber-700 border-amber-200",
        DELETE: "bg-rose-100 text-rose-700 border-rose-200",
      };
      return c[m] || "bg-slate-100 border-slate-200";
    },

    computeBadges() {
      this.badges = [];

      if (route.fetch) this.badges.push("FETCH");
      else if (route.mock?.file) this.badges.push("MOCK_FILE");
      else if (route.mock) this.badges.push("MOCK");

      // Auth badges logic
      if (config.server.auth?.enabled) {
        if (route.auth?.enabled) {
          this.badges.push("AUTH_ROUTE");
        } else {
          this.badges.push("AUTH_GLOBAL");
        }
      }

      // Force route auth badge if explicitly enabled on route
      if (route.auth?.enabled && !this.badges.find((b) => b != "AUTH_ROUTE")) {
        this.badges.push("AUTH_ROUTE");
      }
    },

    /**
     * Constructs the parameters form (Path, Query, Headers, Body).
     * Merges definitions from Config, OpenAPI, and Route settings.
     */
    buildParams() {
      this.params = [];
      const fullPath = config.server.api_prefix + route.path;

      // Path Params from regex
      const pathMatches = route.path.match(/\{([a-zA-Z0-9_]+)\}/g);
      if (pathMatches)
        pathMatches.forEach((p) =>
          this.params.push({
            name: p.replace(/[{}]/g, ""),
            in: "path",
            required: true,
            type: "string",
            value: "",
          })
        );

      // OAS Parameters Integration
      let oasDef = null;
      if (oas.paths) {
        let match = oas.paths[fullPath];
        if (!match) {
          const normalized = fullPath.replace(/:([a-zA-Z0-9_]+)/g, "{$1}");
          match = oas.paths[normalized];
        }
        if (match) oasDef = match[route.method.toLowerCase()];
      }

      if (oasDef && oasDef.parameters)
        oasDef.parameters.forEach((p) => {
          if (!this.params.find((x) => x.name === p.name && x.in === p.in))
            this.params.push({
              name: p.name,
              in: p.in,
              required: p.required || false,
              type: p.schema?.type || "string",
              value: p.example || "",
              example: p.example,
            });
        });

      // Manual Route Query Params
      if (route.query)
        Object.entries(route.query).forEach(([k, def]) => {
          if (!this.params.find((x) => x.name === k && x.in === "query"))
            this.params.push({
              name: k,
              in: "query",
              required: def.required || false,
              type: def.type || "string",
              value: p.example || "",
              example: p.example,
            });
        });

      // Auth Injection into Params List (Visual Only, Real Logic in runRequest)
      if (config.server.auth?.enabled) {
        const authName = config.server.auth.name || "Authorization";
        if (!this.params.find((x) => x.name === authName))
          this.params.push({
            name: authName,
            in: config.server.auth.in || "header",
            required: true,
            type: "string",
            value: "",
          });
      }

      // Body Generation
      if (["POST", "PUT", "PATCH"].includes(route.method)) {
        let bodyObj = route.body_example || {};
        // Fallback to OAS example
        if (
          !Object.keys(bodyObj).length &&
          oasDef?.requestBody?.content?.["application/json"]?.example
        )
          bodyObj = oasDef.requestBody.content["application/json"].example;
        else if (
          !Object.keys(bodyObj).length &&
          oasDef?.requestBody?.content?.["application/json"]?.schema
        ) {
          bodyObj = this.generateSample(
            oasDef.requestBody.content["application/json"].schema
          );
        }
        this.body = JSON.stringify(bodyObj, null, 2);
      }
    },

    // Recursive helper to generate sample JSON from JSON Schema
    generateSample(schema) {
      if (!schema) return null;
      if (schema.type === "object" && schema.properties) {
        const obj = {};
        Object.keys(schema.properties).forEach(
          (k) => (obj[k] = this.generateSample(schema.properties[k]))
        );
        return obj;
      }
      if (schema.type === "array") return [];
      if (schema.type === "string")
        return schema.enum ? schema.enum[0] : "string";
      if (schema.type === "integer" || schema.type === "number") return 0;
      if (schema.type === "boolean") return true;
      return null;
    },

    resetForm() {
      this.result = null;
      this.errors = {};
      this.errorMessage = "";
      this.buildParams();
    },

    /**
     * Executes the API request via browser fetch.
     * Handles parameter injection, validation, and global auth auto-fill.
     */
    async runRequest() {
      this.errors = {};
      this.errorMessage = "";

      //  Validate & Fill Parameters
      let hasError = false;
      this.params.forEach((p) => {
        let val = p.value;

        // Auto-inject global auth values if available and matching location
        if (!val && globalAuthStore[p.name]) {
          const authData = globalAuthStore[p.name];
          if (authData.in === p.in) {
            val = authData.value;
            p.value = val;
          }
        }

        if (p.required && !val) {
          this.errors[p.name] = true;
          hasError = true;
        }
      });
      if (hasError) {
        this.errorMessage = "Missing required fields.";
        return;
      }

      this.reqLoading = true;
      const start = Date.now();

      // Construct URL & Headers
      let url = `http://localhost:${config.server.port}${config.server.api_prefix}${route.path}`;
      const headers = {};
      this.params.forEach((p) => {
        let val = p.value;
        if (
          !val &&
          globalAuthStore[p.name] &&
          globalAuthStore[p.name].in === p.in
        )
          val = globalAuthStore[p.name].value;
        if (val) {
          if (p.in === "path") url = url.replace(`{${p.name}}`, val);
          if (p.in === "query")
            url += (url.includes("?") ? "&" : "?") + `${p.name}=${val}`;
          if (p.in === "header") headers[p.name] = val;
        }
      });

      // Dummy replace for unassigned path params (prevents 404 in simple tests)
      url = url.replace(/\{.*?\}/g, "1");

      const opts = { method: route.method, headers };
      if (["POST", "PUT", "PATCH"].includes(route.method)) {
        opts.headers["Content-Type"] = "application/json";
        try {
          // JSON formatla
          const formattedBody = JSON.stringify(JSON.parse(this.body));
          opts.body = formattedBody;
        } catch (e) {
          opts.body = this.body; // Fallback to raw string if JSON invalid
        }
      }

      // Generate cURL before execution for user reference
      this.reqUrl = url;
      this.generateCurl(url, headers, route.method, opts.body);

      try {
        const res = await fetch(url, opts);

        // Capture Response Headers
        this.resHeaders = {};
        res.headers.forEach((val, key) => {
          this.resHeaders[key] = val;
        });

        // Detect Content Type & Parse Body
        const ct = res.headers.get("content-type") || "";
        let body = null,
          type = "text",
          blobUrl = null;

        if (ct.includes("image")) {
          type = "image";
          const blob = await res.blob();
          blobUrl = URL.createObjectURL(blob);
        } else if (ct.includes("pdf")) {
          type = "pdf";
          const blob = await res.blob();
          blobUrl = URL.createObjectURL(blob);
        } else if (
          ct.includes("json") ||
          ct.includes("text") ||
          ct.includes("xml")
        ) {
          type = "text";
          const txt = await res.text();
          try {
            body = JSON.stringify(JSON.parse(txt), null, 2);
          } catch (e) {
            body = txt;
          }
        } else {
          type = "blob";
        }

        this.result = {
          status: res.status,
          latency: Date.now() - start,
          body,
          type,
          blobUrl,
          contentType: ct.split(";")[0],
        };

        // Syntax Highlight Trigger
        if (type === "text") {
          this.$nextTick(() => {
            this.highlightedBody = Prism.highlight(
              body,
              Prism.languages.json,
              "json"
            );
          });
        }
      } catch (err) {
        this.result = {
          status: "ERR",
          latency: 0,
          body: err.message,
          type: "text",
        };
      } finally {
        this.reqLoading = false;
      }
    },
    copyText(text, type) {
      if (!text) return;
      navigator.clipboard.writeText(text);
      this.activeCopy = type;
      setTimeout(() => {
        if (this.activeCopy === type) {
          this.activeCopy = null;
        }
      }, 2000);
    },
  }));
});
