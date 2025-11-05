# Basic Example

This example demonstrates the fundamental conversion principles:

## Key Conversions

### 1. Query Fields → GET Endpoints
```graphql
users: [User!]!
```
Becomes:
```
GET /users
```

### 2. List Fields → Sub-Resource Endpoints
```graphql
type User {
  posts: [Post!]!
}
```
Becomes:
```
GET /users/{id}/posts
```

### 3. Object References → ID Fields
```graphql
type Post {
  author: User!
}
```
Becomes:
```json
{
  "authorId": "string",
  "description": "Reference to User.id - use GET /users/{authorId}"
}
```

## Benefits

- **No N+1 queries**: Each list requires an explicit endpoint call
- **No depth attacks**: Responses are flat, no arbitrary nesting
- **HTTP cacheable**: Standard GET requests with predictable URLs
- **Type safe**: Inherits GraphQL's type system
