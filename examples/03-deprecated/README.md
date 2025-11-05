# @deprecated Directive Example

This example shows how deprecation information is preserved during conversion.

## GraphQL @deprecated Directive

GraphQL provides the `@deprecated` directive to mark fields and operations that should no longer be used:

```graphql
type Query {
  user(id: ID!): User @deprecated(reason: "Use users query instead")
}

type User {
  emailAddress: String! @deprecated(reason: "Use email field instead")
}
```

## OpenAPI Conversion

The converter maps this to OpenAPI's `deprecated` flag:

### Deprecated Operations
```yaml
paths:
  /user:
    get:
      deprecated: true
      summary: "DEPRECATED: Use users query instead"
```

### Deprecated Fields
```yaml
components:
  schemas:
    User:
      properties:
        emailAddress:
          type: string
          deprecated: true
          description: "DEPRECATED: Use email field instead"
```

## Benefits

1. **API Evolution**: Clearly communicate which parts of the API are being phased out
2. **Tool Support**: Most OpenAPI tools highlight deprecated operations and fields
3. **Documentation**: Reasons for deprecation are preserved in descriptions
4. **Migration Path**: Helps API consumers plan their migration strategy

## Examples in This Schema

- **Deprecated Query**: `user(id: ID!)` → Use `users` query instead
- **Deprecated Mutation**: `registerUser` → Use `createUser` instead
- **Deprecated Fields**:
  - `User.emailAddress` → Use `User.email`
  - `User.username` → Being removed entirely
  - `Post.legacyStatus` → Use extended PostInfo type
