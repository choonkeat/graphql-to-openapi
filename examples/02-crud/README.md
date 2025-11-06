# CRUD / RESTful Pattern Example

This example demonstrates how the converter detects and consolidates REST patterns.

## REST Pattern Detection

### Requirements for Consolidation
To trigger REST pattern consolidation, the schema **must** have:
1. A list query: `{resource}s: [Type!]!` (e.g., `users: [User!]!`)
2. A create mutation: `create{Resource}(...)` (e.g., `createUser(...)`)

### Detected Patterns in This Example

#### User Resource (Full CRUD)
```graphql
Query {
  users: [User!]!          # List
  user(id: ID!): User      # Get
}
Mutation {
  createUser(...)          # Create
  updateUser(id: ID, ...)  # Update
  deleteUser(id: ID)       # Delete
}
```

**Converts to:**
```
GET    /users           → List all users
GET    /users/{id}      → Get user by ID
POST   /users           → Create user
PUT    /users/{id}      → Update user
DELETE /users/{id}      → Delete user
GET    /users/{id}/posts → Get user's posts (sub-resource)
```

#### Post Resource (Full CRUD)
```graphql
Query {
  posts: [Post!]!
  post(id: ID!): Post
}
Mutation {
  createPost(...)
  updatePost(id: ID, ...)
  deletePost(id: ID)
}
```

**Converts to:**
```
GET    /posts       → List all posts
GET    /posts/{id}  → Get post by ID
POST   /posts       → Create post
PUT    /posts/{id}  → Update post
DELETE /posts/{id}  → Delete post
```

#### AuditEntry Resource (Minimal - Only 2 Operations)
```graphql
Query {
  auditEntries: [AuditEntry!]!  # List only
}
Mutation {
  createAuditEntry(...)         # Create only
}
```

**Converts to:**
```
GET    /auditEntries  → List all audit entries
POST   /auditEntries  → Create audit entry
```

This demonstrates the **minimum required** to trigger REST consolidation. Notice:
- No `auditEntry(id)` query → No GET `/auditEntries/{id}` endpoint
- No update/delete mutations → No PUT/DELETE endpoints
- Just the 2 required operations consolidate into REST endpoints

#### Comment (No Pattern - Wrong Prefix)
```graphql
Mutation {
  addComment(...)  # Uses "add" instead of "create"
}
```

**Converts to:**
```
POST /addComment  → Not consolidated (no REST pattern)
```

## Benefits of REST Consolidation

1. **Cleaner API surface**: Standard REST conventions instead of verbose mutations
2. **Better HTTP semantics**: Proper use of GET, POST, PUT, DELETE methods
3. **Idiomatic URLs**: `/users/{id}` instead of `/getUser?id=123`
4. **Automatic sub-resources**: Nested lists become `/users/{id}/posts`
