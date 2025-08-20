# ![MockServer Logo](https://dummyimage.com/600x100/4CAF50/ffffff&text=MockServer)

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/onurartan/mockserver)
[![npm](https://img.shields.io/npm/v/mockserverx)](https://www.npmjs.com/package/mockserverx)
[![License](https://img.shields.io/badge/license-MIT-lightgrey)](https://opensource.org/licenses/MIT)

---

## Table of Contents

- [Overview](#overview)  
- [Purpose](#purpose)  
- [Installation](#installation)  
- [Configuration](#configuration)  
- [Example Usage](#example-usage)  
- [Advanced Features](#advanced-features)  

---

## Overview

**MockServer** is a lightweight, configurable mock API server that allows developers to create realistic API endpoints using a single JSON configuration file. It is designed for both **frontend developers** and **backend engineers** to simulate API responses, test workflows, and prototype integrations without needing a fully functional backend.

Key features include:

- Fully JSON-based configuration  
- Support for REST methods: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`, `HEAD`  
- Mock responses with delay simulation  
- Proxy/fetch real APIs dynamically  
- Authentication support (`apiKey` or `Bearer`)  
- CORS and default headers management  

---

## Purpose

Modern applications rely heavily on APIs. Developing frontend or integration tests often requires a backend that is not fully implemented yet. **MockServer** solves this problem by providing:

- **Fast prototyping:** No need for real backend during early development  
- **Realistic test scenarios:** Support for dynamic response, headers, delays, and authentication  
- **Flexible API simulation:** Fetch real endpoints or mock static JSON files  
- **Single source of truth:** All routes, headers, and responses are managed in one JSON file  

---

## Installation

You can install **MockServer** using **npm**:

```bash
npm install -g mockserverx
```

Start the server:

```bash
mockserver start --config mockserver.json
```

---

## Configuration

**MockServer** relies on a single JSON file, typically named `mockserver.json`, structured according to the [JSON Schema](https://opensource.trymagic.xyz/schemas/mockserver.schema.json). Below is a breakdown of the configuration options:

### Server Configuration

```json
"server": {
  "port": 3000,
  "api_prefix": "/v1",
  "default_headers": {
    "Content-Type": "application/json"
  },
  "default_delay_ms": 0,
  "cors": {
    "enabled": true,
    "allow_origins": ["*"],
    "allow_methods": ["GET","POST","PUT","DELETE"],
    "allow_headers": ["*"],
    "allow_credentials": false
  },
  "auth": {
    "enabled": true,
    "type": "apiKey",
    "name": "apiKey",
    "in": "query",
    "keys": ["secret"]
  }
}
```

**Explanation:**

* `port`: Server listening port (1–65535)
* `api_prefix`: Prefix for all API endpoints
* `default_headers`: Headers applied to every response
* `default_delay_ms`: Global response delay in milliseconds
* `cors`: CORS configuration
* `auth`: Global authentication configuration

---

### Routes

Each route is an object inside the `routes` array:

```json
{
  "name": "Get All Users",
  "tag": "Users",
  "method": "GET",
  "path": "/users",
  "auth": {
    "enabled": true,
    "type": "apiKey",
    "name": "apiKey",
    "in": "query",
    "keys": ["secret"]
  },
  "query": {
    "limit": {
      "type": "integer",
      "description": "Number of records to fetch",
      "required": false
    }
  },
  "fetch": {
    "url": "https://jsonplaceholder.typicode.com/users",
    "headers": {
      "Accept": "application/json"
    }
  }
}
```

**Key properties:**

* `name`: Route display name
* `tag`: Route grouping tag
* `method`: HTTP method (`GET`, `POST`, etc.)
* `path`: Endpoint path, supports path params `{id}`
* `auth`: Optional route-specific authentication
* `query`: Query parameters schema
* `path_params`: Path parameter validation
* `body_schema`: Request body validation (for `POST`, `PUT`, `PATCH`)
* `body_example`: Example request body
* `mock`: Local mock file configuration
* `fetch`: Proxy to real API endpoint

---

## Example Usage

### Start Server

```bash
mockserver start --config mockserver.json
```

### Call Endpoint

```bash
curl http://localhost:3000/v1/users?apiKey=secret
```

### Mock Local JSON

```json
{
  "name": "List Todos (Mock)",
  "tag": "Todos",
  "method": "GET",
  "path": "/todos",
  "mock": {
    "file": "test/mocks/todos.json",
    "status": 200,
    "headers": {
      "X-Source": "MockFile"
    },
    "delay_ms": 50
  }
}
```

---

## Advanced Features

* **Delayed responses** – simulate network latency
* **Dynamic fetch** – route can forward requests to real APIs with path/query substitution
* **Authentication** – supports `apiKey` in query/header or `Bearer` tokens
* **Swagger UI** – generate API docs from `mockserver.json` (default `/docs`)