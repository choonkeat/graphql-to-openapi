# @specifiedBy Directive Example

This example demonstrates how custom scalar types with specification URLs are converted to OpenAPI formats.

## GraphQL Custom Scalars with @specifiedBy

GraphQL allows defining custom scalar types and documenting their format with `@specifiedBy`:

```graphql
scalar UUID @specifiedBy(url: "https://tools.ietf.org/html/rfc4122")
scalar DateTime @specifiedBy(url: "https://scalars.graphql.org/andimarek/date-time")
scalar EmailAddress @specifiedBy(url: "https://tools.ietf.org/html/rfc5322")
```

## OpenAPI Conversion

The converter intelligently maps these to OpenAPI formats and preserves the specification URL:

### UUID Scalar
```yaml
properties:
  id:
    type: string
    format: uuid
    description: "Spec: https://tools.ietf.org/html/rfc4122"
```

### DateTime Scalar
```yaml
properties:
  createdAt:
    type: string
    format: date-time
    description: "Spec: https://scalars.graphql.org/andimarek/date-time"
```

### Custom Scalars (No Recognized Format)
```yaml
properties:
  email:
    type: string
    description: "Spec: https://tools.ietf.org/html/rfc5322"
```

## Format Recognition

The converter automatically recognizes these common specifications:
- **UUID**: Contains "rfc4122" or "uuid" → `format: uuid`
- **DateTime**: Contains "date-time" → `format: date-time`
- **Others**: Preserved as `type: string` with spec URL in description

## Benefits

1. **Type Safety**: Custom scalars map to appropriate OpenAPI formats
2. **Validation**: OpenAPI tools can validate UUID, date-time, etc.
3. **Documentation**: Specification URLs are preserved for reference
4. **Tooling**: Code generators use formats to generate proper types

## Scalars in This Example

| GraphQL Scalar | OpenAPI Format | Specification |
|----------------|----------------|---------------|
| `UUID` | `uuid` | RFC 4122 |
| `DateTime` | `date-time` | ISO 8601 |
| `EmailAddress` | `string` | RFC 5322 |
| `URL` | `string` | RFC 3986 |
| `JSON` | `string` | RFC 7159 |
