# Interfaces & Unions Example

This example demonstrates how GraphQL interfaces and unions are converted to OpenAPI schemas.

## GraphQL Interfaces

Interfaces define a set of fields that implementing types must include. They're useful for creating polymorphic return types while maintaining type safety.

### Single Interface Implementation

**GraphQL:**
```graphql
interface Node {
  id: ID!
}

interface Timestamped {
  createdAt: String!
  updatedAt: String!
}
```

**OpenAPI Conversion:**
Interfaces become regular object schemas with their defined properties:

```yaml
Node:
  type: object
  properties:
    id:
      type: string
  required:
    - id

Timestamped:
  type: object
  properties:
    createdAt:
      type: string
    updatedAt:
      type: string
  required:
    - createdAt
    - updatedAt
```

### Multiple Interface Implementation

**GraphQL:**
```graphql
type Article implements Node & Timestamped & Content {
  id: ID!
  title: String!
  content: String!
  createdAt: String!
  updatedAt: String!
  published: Boolean!
  # ... additional fields
}
```

**OpenAPI Conversion:**
The implementing type includes ALL fields from ALL interfaces plus its own fields:

```yaml
Article:
  type: object
  properties:
    id:
      type: string          # From Node
    createdAt:
      type: string          # From Timestamped
    updatedAt:
      type: string          # From Timestamped
    title:
      type: string          # From Content
    published:
      type: boolean         # From Content
    content:
      type: string          # Article's own field
    # ... more fields
  required:
    - id
    - title
    - content
    - createdAt
    - updatedAt
    - published
```

**Key Points:**
- Interfaces are "flattened" into implementing types
- No inheritance mechanism in OpenAPI (interfaces just ensure field presence)
- Each type is self-contained with all fields

## GraphQL Unions

Unions represent a value that could be one of several types. Unlike interfaces, union types don't share any fields.

### Union Definition

**GraphQL:**
```graphql
union SearchResult = Article | Video | User
```

**OpenAPI Conversion:**
Unions become schemas with `oneOf`:

```yaml
SearchResult:
  oneOf:
    - $ref: '#/components/schemas/Article'
    - $ref: '#/components/schemas/Video'
    - $ref: '#/components/schemas/User'
```

This means: "A SearchResult is exactly one of: Article, Video, or User"

## Query Field Conversions

### Returning Interfaces

**GraphQL:**
```graphql
type Query {
  node(id: ID!): Node
  content: [Content!]!
}
```

**OpenAPI:**
```yaml
/node:
  get:
    parameters:
      - name: id
        in: query
        required: true
    responses:
      '200':
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Node'

/content:
  get:
    responses:
      '200':
        content:
          application/json:
            schema:
              type: array
              items:
                $ref: '#/components/schemas/Content'
```

**Runtime Behavior:**
- The actual response will be a concrete type (Article, Video, etc.)
- The schema describes the guaranteed fields (interface fields)
- Additional fields from concrete types may be present

### Returning Unions

**GraphQL:**
```graphql
type Query {
  search(query: String!): [SearchResult!]!
}
```

**OpenAPI:**
```yaml
/search:
  get:
    parameters:
      - name: query
        in: query
        required: true
    responses:
      '200':
        content:
          application/json:
            schema:
              type: array
              items:
                $ref: '#/components/schemas/SearchResult'
```

**Runtime Behavior:**
- Each array item could be a different type
- Response: `[{...Article...}, {...Video...}, {...User...}]`
- Client must check type to know which fields are available

## Type Discrimination

In GraphQL, clients use `__typename` to discriminate between types:

```graphql
query {
  search(query: "graphql") {
    __typename
    ... on Article {
      title
      content
    }
    ... on Video {
      title
      url
      duration
    }
    ... on User {
      name
      email
    }
  }
}
```

In REST/OpenAPI:
- No built-in type discrimination
- Clients typically check for unique fields
- Or include a `type` field in responses
- Or use different endpoints for different types

## Conversion Summary

| GraphQL Feature | OpenAPI Conversion | Notes |
|----------------|-------------------|-------|
| **Interface Definition** | Object schema | Contains only interface fields |
| **Type implements Interface** | Object schema | Includes all interface fields + own fields |
| **Union Type** | Schema with `oneOf` | Lists all possible types |
| **Query returns Interface** | Reference to interface schema | Guarantees interface fields |
| **Query returns Union** | Reference to union schema | Could be any of the union types |

## Example Conversions

### Example 1: Simple Interface Query

**GraphQL:**
```graphql
query {
  node(id: "123") {
    id
  }
}
```

**REST:**
```bash
GET /node?id=123
```

**Response:**
```json
{
  "id": "123"
}
```
(Could actually be an Article, Video, or User - but only `id` is guaranteed)

### Example 2: Multiple Interfaces

**GraphQL:**
```graphql
query {
  articles {
    id              # From Node
    createdAt       # From Timestamped
    title           # From Content
    content         # Article's own field
  }
}
```

**REST:**
```bash
GET /articles
```

**Response:**
```json
[
  {
    "id": "1",
    "title": "GraphQL Guide",
    "content": "...",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-02T00:00:00Z",
    "published": true,
    "authorId": "user-1",
    "readingTime": 5
  }
]
```

### Example 3: Union Query

**GraphQL:**
```graphql
query {
  search(query: "tutorial") {
    __typename
    ... on Article { title, content }
    ... on Video { title, url, duration }
    ... on User { name, email }
  }
}
```

**REST:**
```bash
GET /search?query=tutorial
```

**Response:**
```json
[
  {
    "id": "1",
    "title": "GraphQL Tutorial",
    "content": "...",
    "published": true,
    "createdAt": "2025-01-01T00:00:00Z"
  },
  {
    "id": "2",
    "title": "Video Tutorial",
    "url": "https://example.com/video",
    "duration": 300,
    "published": true,
    "createdAt": "2025-01-01T00:00:00Z"
  }
]
```
(Mixed types in array - client must inspect fields to determine type)

## Benefits of This Conversion

✅ **Type Safety**: Interface fields are documented and guaranteed
✅ **Flexibility**: Union types support polymorphic responses
✅ **Clear Documentation**: `oneOf` clearly shows possible types
✅ **Standard OpenAPI**: Uses standard OpenAPI 3.0 features

## Limitations

⚠️ **No Native Type Discrimination**: OpenAPI doesn't have `__typename` equivalent
⚠️ **No Fragments**: REST doesn't have GraphQL fragment syntax
⚠️ **Field Selection**: REST endpoints return fixed shapes, not client-selected fields

## Comparison with GraphQL

| Feature | GraphQL | REST/OpenAPI |
|---------|---------|--------------|
| Type Discrimination | `__typename` field | Check unique fields or use `type` property |
| Field Selection | Client chooses fields | Server returns all fields |
| Polymorphic Queries | Inline fragments | oneOf in schema |
| Type Safety | Strong at query time | Strong in schema definition |
| Multiple Interfaces | Supported | Flattened into single type |

## Use Cases

### When Interfaces Work Well:
- Shared fields across types (id, timestamps, etc.)
- Polymorphic collections with common operations
- Base types with variations

### When Unions Work Well:
- Search results across different types
- Notification systems (different event types)
- Activity feeds (mixed content types)
- Error handling (success | error types)

This example provides clearer, simpler demonstrations of interfaces and unions compared to the complex github/starwars schemas!
