# Subscriptions Example

This example demonstrates how GraphQL Subscriptions are converted to Server-Sent Events (SSE) endpoints.

## Why SSE is a Perfect Fit

Server-Sent Events provide:
- **Native browser support**: Built-in `EventSource` API
- **HTTP-based**: Works with standard web infrastructure
- **One-way streaming**: Perfect for GraphQL subscriptions (server → client)
- **Simple protocol**: Plain text event stream format
- **OpenAPI compatible**: Can be documented with `text/event-stream` content type

## Key Conversions

### 1. Subscription Fields → SSE GET Endpoints

**GraphQL:**
```graphql
type Subscription {
  taskUpdated(id: ID!): Task!
}
```

**OpenAPI:**
```yaml
GET /taskUpdated/{id}
  responses:
    '200':
      content:
        text/event-stream:
          schema:
            type: string
```

The endpoint returns an SSE stream that sends `Task` objects whenever the task is updated.

### 2. Subscription Parameters → Path/Query Parameters

**GraphQL:**
```graphql
taskUpdated(id: ID!): Task!
```

**OpenAPI:**
- Required parameters → Path parameters: `/taskUpdated/{id}`
- Optional parameters → Query parameters: `/messageStream?channelId=123&userId=456`

### 3. Parameterless Subscriptions → Simple Endpoints

**GraphQL:**
```graphql
allTasksUpdated: Task!
```

**OpenAPI:**
```
GET /allTasksUpdated
```

### 4. Multiple Parameters Handling

**GraphQL:**
```graphql
messageStream(channelId: ID!, userId: ID): Message!
```

**OpenAPI:**
- Required param → Path: `/messageStream/{channelId}`
- Optional param → Query: `?userId=456`

## SSE Event Format

Each subscription endpoint streams events in this format:

```
event: taskUpdated
data: {"id":"123","title":"Updated task","status":"DONE","createdAt":"2025-01-01T00:00:00Z","updatedAt":"2025-01-01T00:01:00Z"}

event: taskUpdated
data: {"id":"123","title":"Updated task","status":"ARCHIVED","createdAt":"2025-01-01T00:00:00Z","updatedAt":"2025-01-01T00:02:00Z"}
```

## Client Usage Example

**JavaScript (Browser):**
```javascript
const eventSource = new EventSource('/taskUpdated/123');

eventSource.addEventListener('taskUpdated', (event) => {
  const task = JSON.parse(event.data);
  console.log('Task updated:', task);
});

eventSource.addEventListener('error', (error) => {
  console.error('SSE error:', error);
  eventSource.close();
});
```

**cURL:**
```bash
curl -N http://api.example.com/taskUpdated/123
```

## Benefits

✅ **Real-time updates**: Server pushes data to clients as it happens
✅ **Simple protocol**: Plain text format, easy to debug
✅ **Automatic reconnection**: Browsers automatically reconnect on connection loss
✅ **HTTP-based**: Works through proxies, CDNs, and standard infrastructure
✅ **Well-documented**: Fits naturally in OpenAPI specifications
✅ **Widely supported**: Native browser API, libraries available for all platforms

## Conversion Examples

### Example 1: Single Required Parameter

**GraphQL:**
```graphql
subscription($id: ID!) {
  taskUpdated(id: $id) {
    id
    title
    status
  }
}
```

**REST/SSE:**
```bash
GET /taskUpdated/123
Accept: text/event-stream
```

### Example 2: Multiple Parameters

**GraphQL:**
```graphql
subscription($channelId: ID!, $userId: ID) {
  messageStream(channelId: $channelId, userId: $userId) {
    id
    content
    userId
  }
}
```

**REST/SSE:**
```bash
GET /messageStream/channel-456?userId=user-789
Accept: text/event-stream
```

### Example 3: No Parameters

**GraphQL:**
```graphql
subscription {
  allTasksUpdated {
    id
    title
  }
}
```

**REST/SSE:**
```bash
GET /allTasksUpdated
Accept: text/event-stream
```

## Comparison with Other Approaches

| Approach | Real-time | Browser Support | OpenAPI Native | Complexity |
|----------|-----------|-----------------|----------------|------------|
| **SSE** | ✅ | ✅ Native | ✅ | Low |
| WebSocket | ✅ | ✅ | ❌ | High |
| Webhooks | ✅ | ❌ Server only | ✅ | Medium |
| Polling | ⚠️ Delayed | ✅ | ✅ | Low |

SSE provides the best balance of real-time capabilities, simplicity, and OpenAPI compatibility.
