# @constraint Directive Example

This example demonstrates how validation constraints are converted from GraphQL to OpenAPI validation properties.

## GraphQL @constraint Directive

The `@constraint` directive is a de facto standard in the GraphQL community for adding validation rules:

```graphql
type User {
  username: String! @constraint(minLength: 3, maxLength: 20, pattern: "^[a-zA-Z0-9_]+$")
  email: String! @constraint(format: "email")
  age: Int @constraint(min: 13, max: 120)
}
```

## OpenAPI Conversion

The converter maps constraint arguments directly to OpenAPI validation properties:

### String Length Constraints
```yaml
username:
  type: string
  minLength: 3
  maxLength: 20
  pattern: "^[a-zA-Z0-9_]+$"
```

### Numeric Range Constraints
```yaml
age:
  type: integer
  minimum: 13
  maximum: 120

price:
  type: number
  minimum: 0.01
```

### Format Constraints
```yaml
email:
  type: string
  format: email

website:
  type: string
  format: url
```

## Supported Constraint Arguments

| GraphQL Constraint | OpenAPI Property | Applies To | Example |
|-------------------|------------------|------------|---------|
| `minLength` | `minLength` | String | `minLength: 3` |
| `maxLength` | `maxLength` | String | `maxLength: 100` |
| `min` | `minimum` | Number, Integer | `min: 0` |
| `max` | `maximum` | Number, Integer | `max: 100` |
| `pattern` | `pattern` | String | `pattern: "^[A-Z]+$"` |
| `format` | `format` | String | `format: "email"` |

## Real-World Examples

### Username Validation
```graphql
username: String! @constraint(
  minLength: 3,
  maxLength: 20,
  pattern: "^[a-zA-Z0-9_]+$"
)
```
Ensures usernames are 3-20 characters, alphanumeric with underscores.

### Password Strength
```graphql
password: String! @constraint(
  minLength: 8,
  maxLength: 128,
  pattern: "^(?=.*[A-Za-z])(?=.*\\d).+$"
)
```
Requires 8+ characters with at least one letter and one number.

### SKU Format
```graphql
sku: String! @constraint(pattern: "^[A-Z]{3}-[0-9]{5}$")
```
Enforces format: ABC-12345

### Price Range
```graphql
price: Float! @constraint(min: 0.01, max: 1000000)
```
Must be between $0.01 and $1,000,000.

### Discount Percentage
```graphql
discountPercent: Float @constraint(min: 0, max: 100)
```
Valid percentage range: 0-100.

## Benefits

1. **Client-Side Validation**: OpenAPI tools can validate before sending requests
2. **Server-Side Validation**: Same constraints can be enforced on the server
3. **Documentation**: Constraints are visible in generated API docs
4. **Code Generation**: Stricter types in generated client code
5. **Testing**: Clear validation rules for test case generation
