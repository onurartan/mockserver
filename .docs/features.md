# MockServer Features

## 1. Mock Data Routing

MockServer allows defining mock endpoints with JSON data responses. Each route can be configured to return static mock data or proxy requests to external services. Key features:

* **Mock Handlers**: Serve JSON files for predefined routes.
* **Fetch Handlers**: Proxy requests to external URLs while optionally modifying headers, query parameters, or response codes.
* **Dynamic Path Parameters**: Routes can include parameters like `/users/{id}`, which are automatically extracted and used in filtering or proxying.

## 2. Mock Data Filtering

MockServer includes advanced filtering for mock responses. Supported query parameters allow filtering, sorting, and pagination of JSON arrays.

### Filtering Features:

1. **Exact Match Filtering**

   * Use standard query parameters to filter fields exactly:

     ```
     ?status=active
     ```
   * Supports `string`, `number`, and `boolean` fields.

2. **"Like" Filtering**

   * Supports partial matches using the `_like` suffix:

     ```
     ?name_like=john
     ```
   * Case-insensitive substring matching.

3. **Sorting**

   * Sort results by a specific field with `_sort` and `_order` parameters:

     ```
     ?_sort=age&_order=desc
     ```
   * Sorting works on numeric, string, and boolean fields.

4. **Pagination**

   * Paginate large datasets with `_page` and `_limit` parameters:

     ```
     ?_page=2&_limit=10
     ```
   * Supports safe defaults when values are missing or invalid.

The filtering pipeline is executed in the following order: **exact match → like match → sorting → pagination**.

This makes it easy to test frontend applications with complex datasets without touching a real database.

## 3. Authentication Support

MockServer supports authentication both globally and per-route. Two main methods are supported:

1. **API Key Authentication**

   * Can be provided in request headers or query parameters.
   * Example:

     ```http
     GET /users?apiKey=12345
     ```
   * The server validates against a predefined list of keys.

2. **Bearer Token Authentication**

   * Provided in the `Authorization` header:

     ```
     Authorization: Bearer <token>
     ```
   * Only valid tokens in the configuration are accepted.

3. **Flexible Configuration**

   * Route-level authentication can override global server auth.
   * Missing or invalid credentials return proper HTTP status codes (`401 Unauthorized`).

Authentication ensures that mock routes can simulate real-world security behavior for testing.

## 4. Delays and Response Control

* **Configurable Delay**: MockServer can delay responses to simulate network latency.
* **Custom Status Codes**: Routes can be configured to return any HTTP status code.
* **Custom Headers**: Add default or route-specific headers in responses.

---

**Summary:**
MockServer is a flexible tool for creating a complete mock API environment with support for:

* JSON filtering, sorting, and pagination
* Authentication (API keys & Bearer tokens)
* Configurable delays, headers, and status codes
* Dynamic routes with path parameters
* External request proxying

This allows frontend and backend teams to develop and test applications without relying on live APIs.
