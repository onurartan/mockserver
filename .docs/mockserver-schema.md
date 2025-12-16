# MockServer Schema Documentation

![MockServer Logo](https://dummyimage.com/600x100/4CAF50/ffffff\&text=MockServer+Schema)

[![Version](https://img.shields.io/badge/version-0.0.11-blue)](https://github.com/onurartan/mockserver)


## Table of Contents

* [Overview](#overview)
* [Schema Purpose](#schema-purpose)
* [Server Configuration](#server-configuration)
* [Route Configuration](#route-configuration)
* [Groups](#groups)
* [Authentication](#authentication)
* [Mock Responses](#mock-responses)
* [Fetch / Proxy Configuration](#fetch--proxy-configuration)
* [Example Configuration](#example-configuration)
* [Advanced Usage Scenarios](#advanced-usage-scenarios)

---

## Overview

The **MockServer JSON Schema** defines the structure and validation rules for the `mockserver.json` configuration file. This schema ensures consistent, predictable configuration while enabling IDE autocomplete, validation, and documentation generation.

This schema supports:

* Defining **server-wide settings** (port, default headers, CORS, authentication).
* Defining **groups** for route organization.
* Defining **routes** with request validation, authentication, mock responses, and real API fetch/proxy.
* Param validation (`query`, `path`, `headers`) and example request bodies.

---

## Schema Purpose

The schema provides a formal specification for MockServer configuration files, allowing developers to:

* **Validate configuration** before starting the server.
* **Ensure consistency** across multiple environments or projects.
* **Support dynamic API mocking** with structured input for routes, parameters, authentication, and fetch options.

By using the schema, teams can reliably simulate API behavior without requiring a real backend.

---

## Server Configuration

The `server` object defines global settings for the MockServer instance.

```json
"server": {
  "port": 3000,
  "api_prefix": "/v1",
  "default_headers": {
    "Content-Type": "application/json"
  },
  "default_delay_ms": 0,
  "swagger_ui_path": "/docs",
  "cors": {
    "enabled": true,
    "allow_origins": ["*"],
    "allow_methods": ["GET","POST","PUT","DELETE","PATCH","OPTIONS"],
    "allow_headers": ["*"],
    "allow_credentials": false
  },
  "auth": {
    "enabled": true,
    "type": "apiKey",
    "in": "query",
    "name": "apiKey",
    "keys": ["secret"]
  }
}
```

### Key Fields

* `port` — Server listening port (1–65535). Default: 3000.
* `api_prefix` — Prefix for all routes. Default: `/v1`.
* `default_headers` — Applied to all responses unless overridden in routes.
* `default_delay_ms` — Global artificial delay in milliseconds. Default: 0.
* `swagger_ui_path` — Path for auto-generated API documentation. Default: `/docs`.
* `cors` — Configures Cross-Origin Resource Sharing.
* `auth` — Global authentication settings (`apiKey` or `Bearer`).

---

## Groups

Groups are optional and used to organize routes:

```json
"groups": [
  {
    "name": "Users",
    "description": "Endpoints related to user management"
  }
]
```

* `name` — Unique group name (required).
* `description` — Optional description of the group.
* Groups help categorize routes in documentation or Swagger UI.

---

## Route Configuration

Routes define individual API endpoints.

```json
{
  "name": "Get Users",
  "tag": "Users",
  "method": "GET",
  "path": "/users",
  "status": 200,
  "headers": {
    "Cache-Control": "no-cache"
  },
  "delay_ms": 50,
  "query": {
    "limit": { "type": "integer", "required": false, "example": 10 }
  },
  "body_schema": {},
  "body_example": [],
  "auth": { "enabled": true, "type": "apiKey", "in": "query", "name": "apiKey", "keys": ["secret"] }
}
```

### Key Fields

* `name` — Human-readable route name.
* `tag` — Grouping tag for documentation.
* `method` — HTTP method (`GET`, `POST`, etc.).
* `path` — Endpoint path. Supports `{param}` for path parameters.
* `status` — Response status code.
* `headers` — Custom response headers.
* `delay_ms` — Response delay for simulating latency.
* `query` / `path_params` / `request_headers` — Schema for input validation.
* `body_schema` — JSON schema for request body.
* `body_example` — Example body for documentation or mock responses.
* `auth` — Optional route-specific authentication.

---

## Authentication

MockServer supports two authentication types:

1. **API Key (`apiKey`)** – Passed in query or header.
2. **Bearer Token (`bearer`)** – Passed in Authorization header.

```json
"auth": {
  "enabled": true,
  "type": "apiKey",
  "in": "header",
  "name": "X-API-KEY",
  "keys": ["secret1", "secret2"]
}
```

---

## Mock Responses

Local mock files allow returning static responses:

```json
"mock": {
  "file": "mocks/users.json",
  "status": 200,
  "headers": { "X-Source": "MockFile" },
  "delay_ms": 100
}
```

* `file` — Path to local JSON mock file.
* `status` — HTTP response code.
* `headers` — Custom headers.
* `delay_ms` — Simulate response delay.

---

## Fetch / Proxy Configuration

Routes can forward requests to real APIs:

```json
"fetch": {
  "url": "https://api.example.com/users",
  "method": "GET",
  "headers": { "Accept": "application/json" },
  "pass_status": true,
  "delay_ms": 50,
  "timeout_ms": 5000
}
```

* `url` — Target API URL. Supports `{param}` substitution.
* `method` — HTTP method for proxying.
* `headers` — Optional headers for the request.
* `pass_status` — Forward status code from fetched response.
* `delay_ms` — Artificial delay before responding.
* `timeout_ms` — Maximum wait time for real API response.

---

## Example Configuration

```json
{
  "server": { "port": 3000, "api_prefix": "/v1" },
  "routes": [
    {
      "name": "List Todos",
      "method": "GET",
      "path": "/todos",
      "mock": { "file": "mocks/todos.json", "status": 200 }
    },
    {
      "name": "Fetch Users",
      "method": "GET",
      "path": "/users",
      "fetch": { "url": "https://jsonplaceholder.typicode.com/users" }
    }
  ]
}
```

---

## Advanced Usage Scenarios

1. **Simulate network latency** with `delay_ms`.
2. **Dynamic parameter substitution** in fetch URLs.
3. **Route-specific authentication** overriding global auth.
4. **Mock JSON files for offline development or testing**.
5. **Organize routes** using `groups` and `tags`.
6. **Swagger UI generation** for API documentation (`swagger_ui_path`).



## 1. Server Configuration (Global)

### Example JSON

```json
{
  "server": {
    "port": 3000,
    "api_prefix": "/v1",
    "default_headers": {
      "Content-Type": "application/json"
    },
    "default_delay_ms": 100,
    "swagger_ui_path": "/docs",
    "cors": {
      "enabled": true,
      "allow_origins": ["*"],
      "allow_methods": ["GET","POST","PUT","DELETE","PATCH","OPTIONS"],
      "allow_headers": ["*"],
      "allow_credentials": true
    },
    "auth": {
      "enabled": true,
      "type": "apiKey",
      "in": "query",
      "name": "apiKey",
      "keys": ["supersecret"]
    }
  }
}
```

### Explanation

| Field              | Type   | Description               | Example                              |
| ------------------ | ------ | ------------------------- | ------------------------------------ |
| port               | number | Port where server listens | 3000                                 |
| api\_prefix        | string | Prefix for all routes     | "/v1"                                |
| default\_headers   | object | Headers applied globally  | {"Content-Type": "application/json"} |
| default\_delay\_ms | number | Artificial delay (ms)     | 100                                  |
| swagger\_ui\_path  | string | Path for auto-docs        | "/docs"                              |
| cors               | object | CORS configuration        | See JSON example                     |
| auth               | object | Global authentication     | API key in query                     |

**Scenario:**

> You want all endpoints to respond with JSON, support CORS for all domains, and require an API key globally.

---

## 2. Groups

### Example JSON

```json
"groups": [
  {
    "name": "Users",
    "description": "Endpoints for user management"
  },
  {
    "name": "Todos",
    "description": "Endpoints for todo items"
  }
]
```

**Scenario:**

> Organize your endpoints into logical sections for Swagger UI or internal documentation.

---

## 3. Route Configuration

### Example JSON (Mock Response)

```json
{
  "routes": [
    {
      "name": "List Users",
      "tag": "Users",
      "method": "GET",
      "path": "/users",
      "status": 200,
      "headers": {"Cache-Control": "no-cache"},
      "delay_ms": 50,
      "query": {
        "limit": {"type": "integer", "required": false, "example": 5}
      },
      "body_example": [
        {"id": 1, "name": "Alice"},
        {"id": 2, "name": "Bob"}
      ]
    }
  ]
}
```

**Scenario:**

> Return a mock list of users with optional `limit` query parameter. Includes artificial delay to simulate network latency.

---

### Example JSON (Fetch / Proxy Response)

```json
{
  "routes": [
    {
      "name": "Fetch Todos",
      "tag": "Todos",
      "method": "GET",
      "path": "/todos",
      "fetch": {
        "url": "https://jsonplaceholder.typicode.com/todos",
        "method": "GET",
        "headers": {"Accept": "application/json"},
        "pass_status": true,
        "delay_ms": 100,
        "timeout_ms": 5000
      }
    }
  ]
}
```

**Scenario:**

> Proxy real API and optionally add delay. Useful for testing frontend against live data without changing the API calls.

---

## 4. Authentication Examples

### API Key in Header

```json
"auth": {
  "enabled": true,
  "type": "apiKey",
  "in": "header",
  "name": "X-API-KEY",
  "keys": ["secret123"]
}
```

**Scenario:**

> Only requests with correct `X-API-KEY` header will succeed.

### Bearer Token

```json
"auth": {
  "enabled": true,
  "type": "bearer",
  "keys": ["token123", "token456"]
}
```

**Scenario:**

> Accept multiple bearer tokens for testing environments.

---

## 5. Advanced Usage

* **Dynamic Path Params:** `/users/{id}` – Replace `{id}` from query or fetch.
* **Route-specific Authentication:** Override global auth with per-route auth.
* **Multiple Mock Files:** Serve different responses depending on query parameters.
* **Swagger UI Generation:** Automatically document all routes under `/docs`.

---

## 6. Full Example MockServer Config

```json
{
  "server": {
    "port": 3000,
    "api_prefix": "/v1",
    "default_headers": {"Content-Type": "application/json"},
    "cors": {"enabled": true, "allow_origins": ["*"]}
  },
  "groups": [
    {"name": "Users", "description": "User-related endpoints"},
    {"name": "Todos", "description": "Todo-related endpoints"}
  ],
  "routes": [
    {
      "name": "List Users",
      "tag": "Users",
      "method": "GET",
      "path": "/users",
      "status": 200,
      "body_example": [{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]
    },
    {
      "name": "Fetch Todos",
      "tag": "Todos",
      "method": "GET",
      "path": "/todos",
      "fetch": {"url":"https://jsonplaceholder.typicode.com/todos"}
    }
  ]
}
```