# **Handling State Conflicts & Logic Fallbacks**

This guide provides technical insight into MockServer's state management engine and rule evaluation logic. It specifically addresses `409 Conflict` errors and unexpected `null` responses when using conditional cases.

## **1. Understanding `STATE_CONFLICT` (409)**

### **The Mechanism**

MockServer's stateful engine enforces **Primary Key Integrity** on collections. When a route is configured with `action: create`, the engine attempts to insert a new record into the in-memory store.

If the provided `id_field` matches an existing key in the collection, the operation is blocked to prevent accidental data overwrites. This is **intended behavior**, ensuring data consistency during testing.

### **Error Payload**

```json
{
  "error": {
    "code": "STATE_CONFLICT",
    "collection": "payments",
    "message": "Item already exists",
    "id": "ORD-12345"
  }
}

```

### **Resolution Strategies**

* **Unique Identifiers:** Ensure your test suite generates unique IDs (e.g., UUIDs) for every `create` request.
* **Idempotency (Update Action):** If your test logic requires re-running requests with the same ID, change the route action to `update`. This will modify the existing record instead of attempting a duplicate insertion.
* **State Reset:** Utilize the Console Dashboard to clear the collection state between test runs.

---

## **2. The "Null Body" Logic Trap**

### **The Issue**

A request matches a defined `case` condition, but the server returns a `null` or empty body, even though a global `mock.body` is defined.

### **The Logic Flow**

The MockServer engine evaluates responses in a strict priority order:

1. **Cases (`when` match):** Highest Priority.
2. **Stateful Mock:** Standard Priority.
3. **Static Mock:** Fallback Priority.

**Critical Rule:** If a `case` condition evaluates to `true`, the engine **exclusively** executes that case's `then` block. It does **not** fall back to the global `mock` block for missing fields.

### **Incorrect Configuration (The Anti-Pattern)**

In this example, if `amount <= 1000`, the server returns status `200` but an **empty body** because no body is defined in the `then` block.

```yaml
mock:
  status: 201
  body: { "status": "default_success" } # <--- This is IGNORED if a case matches!

cases:
  - when: "request.body.amount <= 1000"
    then:
      status: 200
      # MISSING BODY: Engine returns null.

```

### **Correct Configuration (The Solution)**

You must explicitly define the body within the `case` or remove the case to let the default mock handle it.

```yaml
cases:
  - when: "request.body.amount <= 1000"
    then:
      status: 200
      body: { "status": "specific_success", "amount": "{{request.body.amount}}" }

```

> **Pro Tip:** Use `cases` only when the response structure or data specifically needs to diverge from the default `mock` response. For simple status code changes, ensure the body is also replicated or templated.
