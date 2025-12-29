# MockServer Schema Documentation v1.0.0

[![Version](https://img.shields.io/badge/version-1.0.0-blue)](https://github.com/onurartan/mockserver)
[![Schema](https://img.shields.io/badge/schema-JSON%2FYAML-green)](https://opensource.trymagic.xyz/schemas/mockserver.schema.json)

---

## Table of Contents

- [Overview](#overview)
- [Root Configuration](#root-configuration)
- [Server Configuration](#server-configuration)
- [Console Configuration](#console-configuration)
- [Debug Configuration](#debug-configuration)
- [CORS Configuration](#cors-configuration)
- [Authentication Configuration](#authentication-configuration)
- [Groups Configuration](#groups-configuration)
- [Routes Configuration](#routes-configuration)
- [Parameter Definitions](#parameter-definitions)
- [JSON Schema Validation](#json-schema-validation)
- [Mock Configuration](#mock-configuration)
- [Fetch Configuration](#fetch-configuration)
- [Stateful Configuration](#stateful-configuration)
- [Cases Configuration](#cases-configuration)
- [Template Engine](#template-engine)
- [Complete Examples](#complete-examples)

---

## Overview

MockServer uses a comprehensive JSON/YAML schema to define API mock behavior. This schema enables:

- **Server Configuration**: Port, prefixes, headers, delays
- **Route Definitions**: Endpoints with validation, mocking, and proxying
- **Authentication**: API keys and bearer tokens
- **State Management**: In-memory CRUD operations
- **Conditional Logic**: Dynamic responses based on request data
- **Template Processing**: Dynamic content generation

---

## Root Configuration

The root configuration object contains all MockServer settings.

### JSON Structure
```json
{
  "$schema": "https://opensource.trymagic.xyz/schemas/mockserver.schema.json",
  "server": { /* Server configuration */ },
  "groups": [ /* Optional route groups */ ],
  "routes": [ /* Route definitions */ ]
}
```

### YAML Structure
```yaml
$schema: "https://opensource.trymagic.xyz/schemas/mockserver.schema.json"
server:
  # Server configuration
groups:
  # Optional route groups
routes:
  # Route definitions
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `$schema` | string | No | JSON Schema reference for validation |
| `server` | object | Yes | Global server configuration |
| `groups` | array | No | Route organization groups |
| `routes` | array | Yes | API endpoint definitions |

---

## Server Configuration

Defines global server behavior and default settings.

### JSON Example
```json
{
  "server": {
    "port": 5000,
    "api_prefix": "/api/v1",
    "default_headers": {
      "Content-Type": "application/json",
      "X-API-Version": "1.0"
    },
    "default_delay_ms": 100,
    "swagger_ui_path": "/docs",
    "console": {
      "enabled": true,
      "path": "/console",
      "auth": {
        "enabled": true,
        "username": "admin",
        "password": "secret123"
      }
    },
    "debug": {
      "enabled": true,
      "path": "/__debug"
    },
    "cors": {
      "enabled": true,
      "allow_origins": ["*"],
      "allow_methods": ["GET", "POST", "PUT", "DELETE"],
      "allow_headers": ["Content-Type", "Authorization"],
      "allow_credentials": false
    },
    "auth": {
      "enabled": true,
      "type": "apiKey",
      "name": "X-API-Key",
      "in": "header",
      "keys": ["dev-key-123", "test-key-456"]
    }
  }
}
```

### YAML Example
```yaml
server:
  port: 5000
  api_prefix: "/api/v1"
  default_headers:
    Content-Type: "application/json"
    X-API-Version: "1.0"
  default_delay_ms: 100
  swagger_ui_path: "/docs"
  console:
    enabled: true
    path: "/console"
    auth:
      enabled: true
      username: "admin"
      password: "secret123"
  debug:
    enabled: true
    path: "/__debug"
  cors:
    enabled: true
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE"]
    allow_headers: ["Content-Type", "Authorization"]
    allow_credentials: false
  auth:
    enabled: true
    type: "apiKey"
    name: "X-API-Key"
    in: "header"
    keys: ["dev-key-123", "test-key-456"]
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | integer | 5000 | Server listening port (1-65535) |
| `api_prefix` | string | "" | Prefix for all API routes |
| `default_headers` | object | `{"Content-Type": "application/json"}` | Headers applied to all responses |
| `default_delay_ms` | integer | 0 | Global artificial delay in milliseconds |
| `swagger_ui_path` | string | "/docs" | Path for Swagger UI documentation |
| `console` | object | - | Web console configuration |
| `debug` | object | - | Debug endpoints configuration |
| `cors` | object | - | CORS settings |
| `auth` | object | - | Global authentication settings |

---

## Console Configuration

Web-based management interface settings.

### JSON Example
```json
{
  "console": {
    "enabled": true,
    "path": "/console",
    "auth": {
      "enabled": true,
      "username": "admin",
      "password": "secure-password"
    }
  }
}
```

### YAML Example
```yaml
console:
  enabled: true
  path: "/console"
  auth:
    enabled: true
    username: "admin"
    password: "secure-password"
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | true | Enable/disable web console |
| `path` | string | "/console" | Console access path |
| `auth.enabled` | boolean | true | Enable console authentication |
| `auth.username` | string | "admin" | Console login username |
| `auth.password` | string | "123" | Console login password |

---

## Debug Configuration

Debug endpoints for monitoring and troubleshooting.

### JSON Example
```json
{
  "debug": {
    "enabled": true,
    "path": "/__debug"
  }
}
```

### YAML Example
```yaml
debug:
  enabled: true
  path: "/__debug"
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | true | Enable debug endpoints |
| `path` | string | "/__debug" | Debug endpoints base path |

### Available Endpoints

- `/__debug/health` - Server health and statistics
- `/__debug/requests` - Recent request logs

---

## CORS Configuration

Cross-Origin Resource Sharing settings.

### JSON Example
```json
{
  "cors": {
    "enabled": true,
    "allow_origins": ["http://localhost:3000", "https://app.example.com"],
    "allow_methods": ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"],
    "allow_headers": ["Content-Type", "Authorization", "X-API-Key"],
    "allow_credentials": true
  }
}
```

### YAML Example
```yaml
cors:
  enabled: true
  allow_origins: 
    - "http://localhost:3000"
    - "https://app.example.com"
  allow_methods: 
    - "GET"
    - "POST" 
    - "PUT"
    - "DELETE"
    - "PATCH"
    - "OPTIONS"
  allow_headers:
    - "Content-Type"
    - "Authorization"
    - "X-API-Key"
  allow_credentials: true
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | false | Enable CORS support |
| `allow_origins` | array | ["*"] | Allowed origin domains |
| `allow_methods` | array | ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"] | Allowed HTTP methods |
| `allow_headers` | array | ["Origin", "Content-Type", "Accept", "Authorization"] | Allowed request headers |
| `allow_credentials` | boolean | false | Allow credentials in CORS requests |

---

## Authentication Configuration

Global and route-level authentication settings.

### API Key Authentication (Header)

#### JSON Example
```json
{
  "auth": {
    "enabled": true,
    "type": "apiKey",
    "name": "X-API-Key",
    "in": "header",
    "keys": ["secret-key-123", "dev-key-456"]
  }
}
```

#### YAML Example
```yaml
auth:
  enabled: true
  type: "apiKey"
  name: "X-API-Key"
  in: "header"
  keys:
    - "secret-key-123"
    - "dev-key-456"
```

### API Key Authentication (Query)

#### JSON Example
```json
{
  "auth": {
    "enabled": true,
    "type": "apiKey",
    "name": "api_key",
    "in": "query",
    "keys": ["public-key-789"]
  }
}
```

#### YAML Example
```yaml
auth:
  enabled: true
  type: "apiKey"
  name: "api_key"
  in: "query"
  keys:
    - "public-key-789"
```

### Bearer Token Authentication

#### JSON Example
```json
{
  "auth": {
    "enabled": true,
    "type": "bearer",
    "name": "Authorization",
    "in": "header",
    "keys": ["eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", "valid-jwt-token"]
  }
}
```

#### YAML Example
```yaml
auth:
  enabled: true
  type: "bearer"
  name: "Authorization"
  in: "header"
  keys:
    - "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    - "valid-jwt-token"
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `enabled` | boolean | Enable/disable authentication |
| `type` | string | Authentication type: "apiKey" or "bearer" |
| `name` | string | Parameter name (e.g., "X-API-Key", "Authorization") |
| `in` | string | Location: "header" or "query" |
| `keys` | array | List of valid keys/tokens |

---

## Groups Configuration

Organize routes into logical groups for documentation.

### JSON Example
```json
{
  "groups": [
    {
      "name": "Users",
      "description": "User management endpoints"
    },
    {
      "name": "Orders",
      "description": "Order processing and tracking"
    },
    {
      "name": "Payments",
      "description": "Payment processing endpoints"
    }
  ]
}
```

### YAML Example
```yaml
groups:
  - name: "Users"
    description: "User management endpoints"
  - name: "Orders"
    description: "Order processing and tracking"
  - name: "Payments"
    description: "Payment processing endpoints"
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique group name |
| `description` | string | No | Group description |

---

## Routes Configuration

Define individual API endpoints with validation, mocking, and proxying.

### Basic Route Structure

#### JSON Example
```json
{
  "routes": [
    {
      "name": "Get Users",
      "description": "Retrieve list of users with optional filtering",
      "tag": "Users",
      "method": "GET",
      "path": "/users",
      "status": 200,
      "headers": {
        "Cache-Control": "no-cache",
        "X-Total-Count": "100"
      },
      "delay_ms": 50
    }
  ]
}
```

#### YAML Example
```yaml
routes:
  - name: "Get Users"
    description: "Retrieve list of users with optional filtering"
    tag: "Users"
    method: "GET"
    path: "/users"
    status: 200
    headers:
      Cache-Control: "no-cache"
      X-Total-Count: "100"
    delay_ms: 50
```

### Route with Path Parameters

#### JSON Example
```json
{
  "name": "Get User by ID",
  "method": "GET",
  "path": "/users/{id}",
  "path_params": {
    "id": {
      "type": "integer",
      "required": true,
      "description": "User ID",
      "example": 123
    }
  }
}
```

#### YAML Example
```yaml
name: "Get User by ID"
method: "GET"
path: "/users/{id}"
path_params:
  id:
    type: "integer"
    required: true
    description: "User ID"
    example: 123
```

### Route with Query Parameters

#### JSON Example
```json
{
  "name": "Search Users",
  "method": "GET",
  "path": "/users/search",
  "query": {
    "q": {
      "type": "string",
      "required": true,
      "description": "Search query",
      "example": "john"
    },
    "limit": {
      "type": "integer",
      "required": false,
      "description": "Maximum results",
      "example": 10
    },
    "status": {
      "type": "string",
      "required": false,
      "enum": ["active", "inactive", "pending"],
      "description": "User status filter"
    }
  }
}
```

#### YAML Example
```yaml
name: "Search Users"
method: "GET"
path: "/users/search"
query:
  q:
    type: "string"
    required: true
    description: "Search query"
    example: "john"
  limit:
    type: "integer"
    required: false
    description: "Maximum results"
    example: 10
  status:
    type: "string"
    required: false
    enum: ["active", "inactive", "pending"]
    description: "User status filter"
```

### Route with Request Headers

#### JSON Example
```json
{
  "name": "Create User",
  "method": "POST",
  "path": "/users",
  "request_headers": {
    "Content-Type": {
      "type": "string",
      "required": true,
      "description": "Request content type",
      "example": "application/json"
    },
    "X-Client-Version": {
      "type": "string",
      "required": false,
      "description": "Client application version",
      "example": "1.2.3"
    }
  }
}
```

#### YAML Example
```yaml
name: "Create User"
method: "POST"
path: "/users"
request_headers:
  Content-Type:
    type: "string"
    required: true
    description: "Request content type"
    example: "application/json"
  X-Client-Version:
    type: "string"
    required: false
    description: "Client application version"
    example: "1.2.3"
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable route name |
| `description` | string | No | Route description |
| `tag` | string | No | Group tag for organization |
| `method` | string | Yes | HTTP method (GET, POST, PUT, DELETE, PATCH) |
| `path` | string | Yes | Endpoint path (supports {param} syntax) |
| `status` | integer | No | Default HTTP status code |
| `headers` | object | No | Custom response headers |
| `delay_ms` | integer | No | Route-specific delay in milliseconds |
| `path_params` | object | No | Path parameter definitions |
| `query` | object | No | Query parameter definitions |
| `request_headers` | object | No | Expected request header definitions |
| `body_schema` | object | No | JSON schema for request body validation |
| `body_example` | any | No | Example request body |
| `mock` | object | No | Mock response configuration |
| `fetch` | object | No | Proxy/fetch configuration |
| `stateful` | object | No | State management configuration |
| `cases` | array | No | Conditional response logic |
| `default` | object | No | Default response for cases |
| `auth` | object | No | Route-specific authentication override |

---

## Parameter Definitions

Define validation rules for path parameters, query parameters, and request headers.

### JSON Example
```json
{
  "id": {
    "type": "integer",
    "description": "Unique identifier",
    "required": true,
    "example": 123
  },
  "status": {
    "type": "string",
    "description": "Status filter",
    "required": false,
    "enum": ["active", "inactive", "pending"],
    "example": "active"
  }
}
```

### YAML Example
```yaml
id:
  type: "integer"
  description: "Unique identifier"
  required: true
  example: 123
status:
  type: "string"
  description: "Status filter"
  required: false
  enum: ["active", "inactive", "pending"]
  example: "active"
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Data type: "string", "integer", "number", "boolean" |
| `description` | string | Human-readable description |
| `required` | boolean | Whether parameter is mandatory |
| `enum` | array | List of allowed values |
| `example` | any | Example value for documentation |

---

## JSON Schema Validation

Define request body validation using JSON Schema Draft 7.

### Simple Object Schema

#### JSON Example
```json
{
  "body_schema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "minLength": 2,
        "maxLength": 50
      },
      "email": {
        "type": "string",
        "pattern": "^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$"
      },
      "age": {
        "type": "integer",
        "minimum": 18,
        "maximum": 120
      }
    },
    "required": ["name", "email"]
  }
}
```

#### YAML Example
```yaml
body_schema:
  type: "object"
  properties:
    name:
      type: "string"
      minLength: 2
      maxLength: 50
    email:
      type: "string"
      pattern: "^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$"
    age:
      type: "integer"
      minimum: 18
      maximum: 120
  required: ["name", "email"]
```

### Nested Object Schema

#### JSON Example
```json
{
  "body_schema": {
    "type": "object",
    "properties": {
      "user": {
        "type": "object",
        "properties": {
          "profile": {
            "type": "object",
            "properties": {
              "firstName": {"type": "string"},
              "lastName": {"type": "string"}
            },
            "required": ["firstName"]
          }
        }
      },
      "preferences": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "key": {"type": "string"},
            "value": {"type": "string"}
          }
        }
      }
    }
  }
}
```

#### YAML Example
```yaml
body_schema:
  type: "object"
  properties:
    user:
      type: "object"
      properties:
        profile:
          type: "object"
          properties:
            firstName:
              type: "string"
            lastName:
              type: "string"
          required: ["firstName"]
    preferences:
      type: "array"
      items:
        type: "object"
        properties:
          key:
            type: "string"
          value:
            type: "string"
```

### Schema Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Data type: "object", "array", "string", "integer", "number", "boolean" |
| `properties` | object | Object property definitions (for type: "object") |
| `items` | object | Array item schema (for type: "array") |
| `required` | array | List of required property names |
| `enum` | array | List of allowed values |
| `minimum` | number | Minimum numeric value |
| `maximum` | number | Maximum numeric value |
| `minLength` | integer | Minimum string length |
| `maxLength` | integer | Maximum string length |
| `pattern` | string | Regular expression pattern |
| `additionalProperties` | boolean | Allow undefined properties |

---

## Mock Configuration

Return static responses from inline data or files.

### Inline Mock Response

#### JSON Example
```json
{
  "mock": {
    "status": 200,
    "headers": {
      "X-Source": "MockServer",
      "Cache-Control": "max-age=3600"
    },
    "delay_ms": 100,
    "body": {
      "users": [
        {"id": 1, "name": "John Doe", "email": "john@example.com"},
        {"id": 2, "name": "Jane Smith", "email": "jane@example.com"}
      ],
      "total": 2,
      "page": 1
    }
  }
}
```

#### YAML Example
```yaml
mock:
  status: 200
  headers:
    X-Source: "MockServer"
    Cache-Control: "max-age=3600"
  delay_ms: 100
  body:
    users:
      - id: 1
        name: "John Doe"
        email: "john@example.com"
      - id: 2
        name: "Jane Smith"
        email: "jane@example.com"
    total: 2
    page: 1
```

### File-based Mock Response

#### JSON Example
```json
{
  "mock": {
    "file": "mocks/users.json",
    "status": 200,
    "headers": {
      "X-Source": "MockFile"
    },
    "delay_ms": 50
  }
}
```

#### YAML Example
```yaml
mock:
  file: "mocks/users.json"
  status: 200
  headers:
    X-Source: "MockFile"
  delay_ms: 50
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `body` | any | Inline response body (JSON/YAML data) |
| `file` | string | Path to JSON file containing response data |
| `status` | integer | HTTP status code |
| `headers` | object | Custom response headers |
| `delay_ms` | integer | Artificial delay in milliseconds |

---

## Fetch Configuration

Proxy requests to external APIs with optional modifications.

### Basic Fetch Configuration

#### JSON Example
```json
{
  "fetch": {
    "url": "https://jsonplaceholder.typicode.com/users",
    "method": "GET",
    "headers": {
      "Accept": "application/json",
      "User-Agent": "MockServer/1.0"
    },
    "timeout_ms": 5000,
    "delay_ms": 100
  }
}
```

#### YAML Example
```yaml
fetch:
  url: "https://jsonplaceholder.typicode.com/users"
  method: "GET"
  headers:
    Accept: "application/json"
    User-Agent: "MockServer/1.0"
  timeout_ms: 5000
  delay_ms: 100
```

### Fetch with Path Parameters

#### JSON Example
```json
{
  "path": "/users/{id}",
  "fetch": {
    "url": "https://jsonplaceholder.typicode.com/users/{id}",
    "method": "GET"
  }
}
```

#### YAML Example
```yaml
path: "/users/{id}"
fetch:
  url: "https://jsonplaceholder.typicode.com/users/{id}"
  method: "GET"
```

### Fetch with Query Parameters

#### JSON Example
```json
{
  "fetch": {
    "url": "https://api.example.com/search",
    "method": "GET",
    "query_params": {
      "api_key": "your-api-key",
      "version": "v2"
    }
  }
}
```

#### YAML Example
```yaml
fetch:
  url: "https://api.example.com/search"
  method: "GET"
  query_params:
    api_key: "your-api-key"
    version: "v2"
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `url` | string | Target API URL (supports {param} substitution) |
| `method` | string | HTTP method for upstream request |
| `headers` | object | Headers to send with request |
| `query_params` | object | Additional query parameters |
| `pass_status` | boolean | Forward upstream HTTP status code |
| `delay_ms` | integer | Artificial delay before response |
| `timeout_ms` | integer | Request timeout in milliseconds |

---

## Stateful Configuration

Enable in-memory CRUD operations for simulating database behavior.

### Create Operation

#### JSON Example
```json
{
  "name": "Create User",
  "method": "POST",
  "path": "/users",
  "stateful": {
    "collection": "users",
    "action": "create",
    "id_field": "id"
  },
  "body_schema": {
    "type": "object",
    "properties": {
      "id": {"type": "integer"},
      "name": {"type": "string"},
      "email": {"type": "string"}
    },
    "required": ["id", "name", "email"]
  },
  "mock": {
    "status": 201,
    "body": "{{state.created}}"
  }
}
```

#### YAML Example
```yaml
name: "Create User"
method: "POST"
path: "/users"
stateful:
  collection: "users"
  action: "create"
  id_field: "id"
body_schema:
  type: "object"
  properties:
    id:
      type: "integer"
    name:
      type: "string"
    email:
      type: "string"
  required: ["id", "name", "email"]
mock:
  status: 201
  body: "{{state.created}}"
```

### Read Operations

#### JSON Example
```json
{
  "name": "Get User",
  "method": "GET",
  "path": "/users/{id}",
  "stateful": {
    "collection": "users",
    "action": "get",
    "id_field": "id"
  },
  "mock": {
    "status": 200,
    "body": "{{state.item}}"
  }
}
```

#### YAML Example
```yaml
name: "Get User"
method: "GET"
path: "/users/{id}"
stateful:
  collection: "users"
  action: "get"
  id_field: "id"
mock:
  status: 200
  body: "{{state.item}}"
```

### Update Operation

#### JSON Example
```json
{
  "name": "Update User",
  "method": "PUT",
  "path": "/users/{id}",
  "stateful": {
    "collection": "users",
    "action": "update",
    "id_field": "id"
  },
  "mock": {
    "status": 200,
    "body": "{{state.updated}}"
  }
}
```

#### YAML Example
```yaml
name: "Update User"
method: "PUT"
path: "/users/{id}"
stateful:
  collection: "users"
  action: "update"
  id_field: "id"
mock:
  status: 200
  body: "{{state.updated}}"
```

### Delete Operation

#### JSON Example
```json
{
  "name": "Delete User",
  "method": "DELETE",
  "path": "/users/{id}",
  "stateful": {
    "collection": "users",
    "action": "delete",
    "id_field": "id"
  },
  "mock": {
    "status": 200,
    "body": {
      "success": true,
      "message": "User deleted successfully"
    }
  }
}
```

#### YAML Example
```yaml
name: "Delete User"
method: "DELETE"
path: "/users/{id}"
stateful:
  collection: "users"
  action: "delete"
  id_field: "id"
mock:
  status: 200
  body:
    success: true
    message: "User deleted successfully"
```

### List Operation

#### JSON Example
```json
{
  "name": "List Users",
  "method": "GET",
  "path": "/users",
  "stateful": {
    "collection": "users",
    "action": "list"
  },
  "mock": {
    "status": 200,
    "body": "{{state.list}}"
  }
}
```

#### YAML Example
```yaml
name: "List Users"
method: "GET"
path: "/users"
stateful:
  collection: "users"
  action: "list"
mock:
  status: 200
  body: "{{state.list}}"
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `collection` | string | Name of the in-memory collection |
| `action` | string | CRUD operation: "create", "get", "update", "delete", "list" |
| `id_field` | string | Field name used as unique identifier |

### State Template Variables

| Variable | Description |
|----------|-------------|
| `{{state.created}}` | Newly created item (create action) |
| `{{state.item}}` | Retrieved item (get action) |
| `{{state.updated}}` | Updated item (update action) |
| `{{state.list}}` | All items in collection (list action) |

---

## Cases Configuration

Implement conditional logic for dynamic responses based on request data.

### Basic Cases

#### JSON Example
```json
{
  "name": "Payment Processing",
  "method": "POST",
  "path": "/payments",
  "cases": [
    {
      "when": "request.body.amount > 1000",
      "then": {
        "status": 403,
        "body": {
          "error": "Amount exceeds limit",
          "code": "AMOUNT_TOO_HIGH",
          "max_allowed": 1000
        },
        "headers": {
          "X-Error-Code": "AMOUNT_TOO_HIGH"
        },
        "delay_ms": 100
      }
    },
    {
      "when": "request.body.currency != 'USD'",
      "then": {
        "status": 400,
        "body": {
          "error": "Unsupported currency",
          "supported": ["USD"]
        }
      }
    }
  ],
  "default": {
    "status": 200,
    "body": {
      "payment_id": "{{uuid}}",
      "status": "success",
      "processed_at": "{{date}}"
    }
  }
}
```

#### YAML Example
```yaml
name: "Payment Processing"
method: "POST"
path: "/payments"
cases:
  - when: "request.body.amount > 1000"
    then:
      status: 403
      body:
        error: "Amount exceeds limit"
        code: "AMOUNT_TOO_HIGH"
        max_allowed: 1000
      headers:
        X-Error-Code: "AMOUNT_TOO_HIGH"
      delay_ms: 100
  - when: "request.body.currency != 'USD'"
    then:
      status: 400
      body:
        error: "Unsupported currency"
        supported: ["USD"]
default:
  status: 200
  body:
    payment_id: "{{uuid}}"
    status: "success"
    processed_at: "{{date}}"
```

### Complex Conditions

#### JSON Example
```json
{
  "cases": [
    {
      "when": "request.body.priority == 'high' AND request.body.amount <= 5000",
      "then": {
        "status": 200,
        "body": {
          "order_id": "{{request.body.order_id}}",
          "status": "processed",
          "estimated_delivery": "{{dateFuture days=1}}"
        },
        "headers": {
          "X-Priority": "high"
        }
      }
    },
    {
      "when": "request.headers.X-API-Key != 'secret'",
      "then": {
        "status": 401,
        "body": {
          "error": "Invalid API key"
        }
      }
    },
    {
      "when": "type(request.body.user_id) != 'number'",
      "then": {
        "status": 400,
        "body": {
          "error": "user_id must be a number"
        }
      }
    }
  ]
}
```

#### YAML Example
```yaml
cases:
  - when: "request.body.priority == 'high' AND request.body.amount <= 5000"
    then:
      status: 200
      body:
        order_id: "{{request.body.order_id}}"
        status: "processed"
        estimated_delivery: "{{dateFuture days=1}}"
      headers:
        X-Priority: "high"
  - when: "request.headers.X-API-Key != 'secret'"
    then:
      status: 401
      body:
        error: "Invalid API key"
  - when: "type(request.body.user_id) != 'number'"
    then:
      status: 400
      body:
        error: "user_id must be a number"
```

### Condition Syntax

#### Request Data Access
- `request.body.field` - Access request body fields
- `request.query.param` - Access query parameters
- `request.headers.name` - Access request headers
- `request.path.param` - Access path parameters

#### Operators
- `==`, `!=` - Equality comparison
- `>`, `>=`, `<`, `<=` - Numeric comparison
- `AND`, `OR` - Logical operators
- `type(value)` - Type checking function

#### Type Checking
- `type(request.body.field) == 'string'`
- `type(request.body.field) == 'number'`
- `type(request.body.field) == 'boolean'`

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `when` | string | Boolean expression to evaluate |
| `then.status` | integer | HTTP status code |
| `then.body` | any | Response body |
| `then.headers` | object | Response headers |
| `then.delay_ms` | integer | Response delay |

---

## Template Engine

Generate dynamic content using template variables and functions.

### Request Data Templates

#### JSON Example
```json
{
  "body": {
    "user_id": "{{request.body.user_id}}",
    "search_query": "{{request.query.q}}",
    "api_key": "{{request.headers.X-API-Key}}",
    "resource_id": "{{request.path.id}}"
  }
}
```

#### YAML Example
```yaml
body:
  user_id: "{{request.body.user_id}}"
  search_query: "{{request.query.q}}"
  api_key: "{{request.headers.X-API-Key}}"
  resource_id: "{{request.path.id}}"
```

### Generator Functions

#### JSON Example
```json
{
  "body": {
    "id": "{{uuid}}",
    "name": "{{name}}",
    "email": "{{email}}",
    "created_at": "{{date}}",
    "expires_at": "{{dateFuture days=30}}",
    "random_code": "{{number min=1000 max=9999}}",
    "is_active": "{{bool}}"
  }
}
```

#### YAML Example
```yaml
body:
  id: "{{uuid}}"
  name: "{{name}}"
  email: "{{email}}"
  created_at: "{{date}}"
  expires_at: "{{dateFuture days=30}}"
  random_code: "{{number min=1000 max=9999}}"
  is_active: "{{bool}}"
```

### Available Functions

| Function | Description | Example |
|----------|-------------|---------|
| `{{uuid}}` | Generate UUID | `550e8400-e29b-41d4-a716-446655440000` |
| `{{name}}` | Generate random name | `John Doe` |
| `{{email}}` | Generate random email | `john.doe@example.com` |
| `{{date}}` | Current date | `2024-01-15` |
| `{{dateFuture days=N}}` | Future date | `2024-02-15` |
| `{{dateNow}}` | Current date | `2024-01-15` |
| `{{number min=X max=Y}}` | Random number | `1234` |
| `{{bool}}` | Random boolean | `true` |

---

## Complete Examples

### E-commerce API Configuration

#### JSON Example
```json
{
  "$schema": "https://opensource.trymagic.xyz/schemas/mockserver.schema.json",
  "server": {
    "port": 3000,
    "api_prefix": "/api/v1",
    "default_headers": {
      "Content-Type": "application/json"
    },
    "cors": {
      "enabled": true,
      "allow_origins": ["*"]
    },
    "auth": {
      "enabled": true,
      "type": "apiKey",
      "name": "X-API-Key",
      "in": "header",
      "keys": ["secret-key-123"]
    }
  },
  "groups": [
    {"name": "Users", "description": "User management"},
    {"name": "Products", "description": "Product catalog"},
    {"name": "Orders", "description": "Order processing"}
  ],
  "routes": [
    {
      "name": "Create User",
      "tag": "Users",
      "method": "POST",
      "path": "/users",
      "stateful": {
        "collection": "users",
        "action": "create",
        "id_field": "id"
      },
      "body_schema": {
        "type": "object",
        "properties": {
          "id": {"type": "integer"},
          "name": {"type": "string"},
          "email": {"type": "string"}
        },
        "required": ["id", "name", "email"]
      },
      "mock": {
        "status": 201,
        "body": "{{state.created}}"
      }
    },
    {
      "name": "Get Products",
      "tag": "Products",
      "method": "GET",
      "path": "/products",
      "query": {
        "category": {"type": "string", "required": false},
        "limit": {"type": "integer", "required": false}
      },
      "fetch": {
        "url": "https://api.example.com/products",
        "headers": {"Accept": "application/json"}
      }
    },
    {
      "name": "Process Order",
      "tag": "Orders",
      "method": "POST",
      "path": "/orders",
      "body_schema": {
        "type": "object",
        "properties": {
          "amount": {"type": "number"},
          "currency": {"type": "string"}
        },
        "required": ["amount", "currency"]
      },
      "cases": [
        {
          "when": "request.body.amount > 1000",
          "then": {
            "status": 403,
            "body": {"error": "Amount exceeds limit"}
          }
        }
      ],
      "default": {
        "status": 200,
        "body": {
          "order_id": "{{uuid}}",
          "status": "success"
        }
      }
    }
  ]
}
```

#### YAML Example
```yaml
$schema: "https://opensource.trymagic.xyz/schemas/mockserver.schema.json"
server:
  port: 3000
  api_prefix: "/api/v1"
  default_headers:
    Content-Type: "application/json"
  cors:
    enabled: true
    allow_origins: ["*"]
  auth:
    enabled: true
    type: "apiKey"
    name: "X-API-Key"
    in: "header"
    keys: ["secret-key-123"]

groups:
  - name: "Users"
    description: "User management"
  - name: "Products"
    description: "Product catalog"
  - name: "Orders"
    description: "Order processing"

routes:
  - name: "Create User"
    tag: "Users"
    method: "POST"
    path: "/users"
    stateful:
      collection: "users"
      action: "create"
      id_field: "id"
    body_schema:
      type: "object"
      properties:
        id:
          type: "integer"
        name:
          type: "string"
        email:
          type: "string"
      required: ["id", "name", "email"]
    mock:
      status: 201
      body: "{{state.created}}"

  - name: "Get Products"
    tag: "Products"
    method: "GET"
    path: "/products"
    query:
      category:
        type: "string"
        required: false
      limit:
        type: "integer"
        required: false
    fetch:
      url: "https://api.example.com/products"
      headers:
        Accept: "application/json"

  - name: "Process Order"
    tag: "Orders"
    method: "POST"
    path: "/orders"
    body_schema:
      type: "object"
      properties:
        amount:
          type: "number"
        currency:
          type: "string"
      required: ["amount", "currency"]
    cases:
      - when: "request.body.amount > 1000"
        then:
          status: 403
          body:
            error: "Amount exceeds limit"
    default:
      status: 200
      body:
        order_id: "{{uuid}}"
        status: "success"
```

---

This comprehensive schema documentation provides all the necessary information to configure MockServer for any API mocking scenario, from simple static responses to complex stateful applications with conditional logic and external API proxying.