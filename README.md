# GraphQL to OpenAPI Converter

> **GraphQL the Good Parts 🤝 OpenAPI the Good Parts**

A CLI tool that converts GraphQL schemas to OpenAPI 3.0 specifications, solving GraphQL's operational complexity problems (N+1 queries, depth attacks, caching difficulties) while maintaining type safety.

## The Problem

GraphQL schemas are succinct and type-safe, but query flexibility creates operational complexity:

- **N+1 Query Problem**: Nested queries cause cascading database calls
- **Depth/Complexity Attacks**: Malicious deeply nested queries can DOS servers
- **Unpredictable Performance**: Client controls query complexity
- **Caching Difficulty**: POST-based queries harder to cache than REST

## The Solution

This tool converts GraphQL schemas to OpenAPI specs using conventions that **eliminate N+1 and depth attacks by design**:

- Lists → separate endpoints: `/users/{id}/posts`
- Objects → ID references: `authorId`
- No nested embedding, flat response shapes
- Standard HTTP caching with GET requests

## Installation

```bash
# Build from source
make build

# Or build directly
go build -o graphql-to-openapi .
```

## Quick Start

```bash
# Convert a GraphQL schema to OpenAPI
./graphql-to-openapi -schema schema.graphql -output api.yaml
```

**[View Live Examples →](https://graphql-to-openapi.netlify.app)**

## Usage

```bash
graphql-to-openapi [options]

Basic Options:
  -schema string
        GraphQL schema file (required)
  -output string
        Output OpenAPI file (default "openapi.yaml")
  -format string
        Output format: yaml or json (default "yaml")

API Metadata:
  -title string
        API title (default "Converted from GraphQL")
  -version string
        API version (default "1.0.0")
  -base-url string
        Base URL for the API
  -path-prefix string
        Path prefix for all endpoints (e.g., "/api/v1")

REST Pattern Detection:
  -detect-rest-patterns
        Enable REST pattern detection (default true)
  -pluralize-suffixes string
        Custom pluralization suffix rules as JSON file

Advanced Pluralization:
  -pluralize-es-suffixes string
        Suffixes that get 'es' added (default "s,x,z,ch,sh")
  -pluralize-ies-suffix string
        Suffix that triggers 'ies' conversion (default "y")
  -pluralize-default-suffix string
        Default suffix to add (default "s")

Advanced CRUD Prefixes:
  -crud-prefix-create string
        Prefix for create operations (default "create")
  -crud-prefix-update string
        Prefix for update operations (default "update")
  -crud-prefix-delete string
        Prefix for delete operations (default "delete")

Examples:
  # Basic conversion
  graphql-to-openapi -schema schema.graphql -output api.yaml

  # With API metadata
  graphql-to-openapi -schema schema.graphql \
    -title "My API" -version "2.0.0" -path-prefix "/api/v2"

  # Custom CRUD prefixes for non-English APIs
  graphql-to-openapi -schema schema.graphql \
    -crud-prefix-create "add" -crud-prefix-update "modify"
```

## Examples

**[View Live Examples →](https://graphql-to-openapi.netlify.app)**

The examples demonstrate various features:

### 01-basic
Basic conversion showing how queries become GET endpoints, lists become sub-resources, and objects become ID references.

**GraphQL:**
```graphql
type Query {
  users: [User!]!
}

type User {
  id: ID!
  name: String!
  posts: [Post!]!
}
```

**OpenAPI:**
- `GET /users` - List all users
- `GET /users/{id}/posts` - Get user's posts (avoids N+1)

[View Examples →](https://graphql-to-openapi.netlify.app)

### 02-crud
REST pattern detection that consolidates CRUD operations into proper RESTful routes.

**Detection Rule:** Must have both `{resource}s: [T]` **and** `create{Resource}(...)` to trigger.

**GraphQL:**
```graphql
type Query {
  users: [User!]!
  user(id: ID!): User
}

type Mutation {
  createUser(...): User!
  updateUser(id: ID, ...): User!
  deleteUser(id: ID!): Boolean!
}
```

**OpenAPI:**
- `GET /users` - List
- `GET /users/{id}` - Get
- `POST /users` - Create
- `PUT /users/{id}` - Update
- `DELETE /users/{id}` - Delete

[View Examples →](https://graphql-to-openapi.netlify.app)

### 03-deprecated
Demonstrates how `@deprecated` directives are converted to OpenAPI's deprecated flag.

**GraphQL:**
```graphql
type User {
  email: String!
  emailAddress: String! @deprecated(reason: "Use email field instead")
}
```

**OpenAPI:**
```yaml
emailAddress:
  type: string
  deprecated: true
  description: "DEPRECATED: Use email field instead"
```

[View Examples →](https://graphql-to-openapi.netlify.app)

### 04-specifiedby
Shows how custom scalar types with `@specifiedBy` map to OpenAPI formats.

**GraphQL:**
```graphql
scalar UUID @specifiedBy(url: "https://tools.ietf.org/html/rfc4122")
scalar DateTime @specifiedBy(url: "https://scalars.graphql.org/andimarek/date-time")
```

**OpenAPI:**
```yaml
id:
  type: string
  format: uuid
  description: "Spec: https://tools.ietf.org/html/rfc4122"

createdAt:
  type: string
  format: date-time
  description: "Spec: https://scalars.graphql.org/andimarek/date-time"
```

[View Examples →](https://graphql-to-openapi.netlify.app)

### 05-constraint
Validation constraints via `@constraint` directive converted to OpenAPI validation properties.

**GraphQL:**
```graphql
type User {
  username: String! @constraint(minLength: 3, maxLength: 20, pattern: "^[a-zA-Z0-9_]+$")
  email: String! @constraint(format: "email")
  age: Int @constraint(min: 13, max: 120)
}
```

**OpenAPI:**
```yaml
username:
  type: string
  minLength: 3
  maxLength: 20
  pattern: "^[a-zA-Z0-9_]+$"

email:
  type: string
  format: email

age:
  type: integer
  minimum: 13
  maximum: 120
```

[View Examples →](https://graphql-to-openapi.netlify.app)

### 08-subscriptions
GraphQL subscriptions converted to Server-Sent Events (SSE) endpoints for real-time streaming.

**GraphQL:**
```graphql
type Subscription {
  taskUpdated(id: ID!): Task!
  newMessage(channelId: ID!): Message!
  allTasksUpdated: Task!
}
```

**OpenAPI:**
- `GET /taskUpdated/{id}` - SSE stream with path parameter
- `GET /newMessage/{channelId}` - SSE stream for channel messages
- `GET /allTasksUpdated` - SSE stream for all tasks

Each endpoint returns `text/event-stream` content type with events formatted as:
```
event: taskUpdated
data: {"id":"123","title":"Task name",...}
```

**Benefits:**
- Native browser support via `EventSource` API
- HTTP-based, works with standard infrastructure
- Perfect fit for OpenAPI documentation
- Real-time updates with automatic reconnection

[View Examples →](https://graphql-to-openapi.netlify.app)

### 09-interfaces
GraphQL interfaces and unions converted to OpenAPI schemas with proper type representation.

**GraphQL:**
```graphql
interface Node {
  id: ID!
}

interface Timestamped {
  createdAt: String!
  updatedAt: String!
}

type Article implements Node & Timestamped & Content {
  id: ID!
  title: String!
  content: String!
  createdAt: String!
  updatedAt: String!
  published: Boolean!
}

union SearchResult = Article | Video | User
```

**OpenAPI:**
- Interfaces → Object schemas with interface fields
- Implementing types → Object schemas with ALL fields (interface + own)
- Unions → Schemas with `oneOf` listing possible types

```yaml
# Interface becomes object schema
Node:
  type: object
  properties:
    id:
      type: string

# Implementing type gets all fields
Article:
  type: object
  properties:
    id: ...          # from Node
    createdAt: ...   # from Timestamped
    updatedAt: ...   # from Timestamped
    title: ...       # from Content
    published: ...   # from Content
    content: ...     # Article's own

# Union becomes oneOf
SearchResult:
  oneOf:
    - $ref: '#/components/schemas/Article'
    - $ref: '#/components/schemas/Video'
    - $ref: '#/components/schemas/User'
```

**Benefits:**
- Clear polymorphic type representation
- Standard OpenAPI `oneOf` for unions
- Interface fields guaranteed in implementing types
- Simple, understandable examples (vs complex github/starwars schemas)

[View Examples →](https://graphql-to-openapi.netlify.app)

## Conversion Rules

### Phase 1: REST Pattern Detection (Optional)

When enabled (default), detects and consolidates REST patterns:

```
GraphQL                          OpenAPI
─────────────────────────────────────────────────
users: [User!]!           →      GET /users
user(id: ID!): User       →      GET /users/{id}
createUser(...)           →      POST /users
updateUser(id: ID, ...)   →      PUT /users/{id}
deleteUser(id: ID)        →      DELETE /users/{id}
```

**Requirements:**
- Must have both `{resource}s: [T]` and `create{Resource}(...)`
- Strict name matching only (no fuzzy matching)

### Phase 2: Simple 1:1 Mapping

Fields not matching REST patterns:

```
GraphQL Query Field           OpenAPI
──────────────────────────────────────────────
searchPosts(q: String)   →    GET /searchPosts?q=hello
calculateTotal(...)      →    GET /calculateTotal

GraphQL Mutation Field        OpenAPI
──────────────────────────────────────────────
approveOrder(id: ID)     →    POST /approveOrder
```

### List Fields → Sub-Resources

```
GraphQL Type Field            OpenAPI
──────────────────────────────────────────────
User.posts: [Post!]!     →    GET /users/{id}/posts
```

### Object References → ID Fields

```
GraphQL Type Field            OpenAPI Schema
──────────────────────────────────────────────
Post.author: User!       →    Post.authorId: string
```

### Subscriptions → SSE Endpoints

GraphQL subscriptions are converted to Server-Sent Events (SSE) endpoints:

```
GraphQL Subscription                    OpenAPI
──────────────────────────────────────────────────────────────────
taskUpdated(id: ID!): Task!       →     GET /taskUpdated/{id}
                                        Content-Type: text/event-stream

allTasksUpdated: Task!            →     GET /allTasksUpdated
                                        Content-Type: text/event-stream

messageStream(                    →     GET /messageStream/{channelId}?userId=123
  channelId: ID!,                       Content-Type: text/event-stream
  userId: ID
): Message!
```

**Parameter Handling:**
- First required parameter → path parameter
- Other parameters → query parameters
- No parameters → simple path

**SSE Event Format:**
```
event: taskUpdated
data: {"id":"123","title":"Updated task",...}
```

## Benefits

✅ **No N+1 Queries**: Each list requires explicit endpoint call
✅ **No Depth Attacks**: Flat responses, no arbitrary nesting
✅ **HTTP Cacheable**: Standard GET requests with predictable URLs
✅ **Type Safe**: Inherits GraphQL's excellent type system
✅ **Simpler**: REST is easier to understand and debug
✅ **Better Tooling**: Massive OpenAPI ecosystem

## Makefile Targets

```bash
make build      # Build the CLI binary
make clean      # Remove build artifacts
make test       # Run tests
make examples   # Generate OpenAPI files from all examples
make docs       # Generate HTML documentation from OpenAPI files
make all        # Build everything and generate docs
```

## Documentation

**[View Live Documentation →](https://graphql-to-openapi.netlify.app)**

Each example includes:
- 📝 GraphQL Schema (syntax highlighted)
- 📄 OpenAPI YAML (syntax highlighted)
- 📖 OpenAPI Docs (interactive HTML)

To generate documentation locally:
```bash
make docs
open examples/index.html
```

## Project Structure

```
.
├── main.go                    # CLI entry point
├── converter/                 # Core conversion logic
│   ├── converter.go           # Main converter implementation
│   ├── types.go               # OpenAPI type definitions
│   └── yaml.go                # YAML marshaling
├── examples/                  # Example schemas with generated documentation
│   ├── 01-basic/
│   │   ├── schema.graphql     # Source GraphQL schema
│   │   └── openapi.yaml       # Generated OpenAPI spec
│   ├── 02-crud/
│   ├── ...
│   ├── index.html             # Documentation index
│   ├── *.schema.html          # Syntax-highlighted GraphQL schemas
│   ├── *.yaml.html            # Syntax-highlighted OpenAPI YAML
│   └── *.redoc.html           # Interactive OpenAPI documentation
├── scripts/                   # Build and generation scripts
│   ├── generate-docs.sh       # Generate all HTML documentation
│   └── generate-index.sh      # Generate documentation index page
└── Makefile                   # Build automation
```

## Dependencies

This project uses minimal dependencies:

- `github.com/vektah/gqlparser/v2` - GraphQL schema parsing
- `gopkg.in/yaml.v3` - YAML output (OpenAPI standard format)

For documentation generation (optional):
- `redoc-cli` - OpenAPI HTML documentation generator (via npm)

## License

MIT License - See [LICENSE](LICENSE) file for details.
