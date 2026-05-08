# Zaptools-go Technical Documentation

## Overview
`zaptools-go` is an Event-Driven WebSocket management toolkit for Go. It abstracts the underlying WebSocket implementation (such as Gorilla WebSocket, Echo, or Fiber) via a standard interface, allowing developers to build robust, event-based real-time applications similar to Socket.io.

## Architecture

### 1. `StdConn` Interface
The core abstraction of the library. Any WebSocket framework that implements this interface can be used seamlessly with Zaptools.
```go
type StdConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}
```

### 2. Event Registration (`EventRegister` & `EventBook`)
- **`EventBook`**: A thread-safe concurrent map (`sync.RWMutex`) that safely stores registered events.
- **`EventRegister`**: Exposes the `OnEvent` method to bind string event names to callback functions (`func(ctx *EventContext)`). 

### 3. Connection Lifecycle (`ZapConnector` & `EventProcessor`)
When a new WebSocket upgrade occurs, a `ZapConnector` is created with `ConnectorOptions`. Calling `Start()` initializes an `EventProcessor` which handles the lifecycle:
1. Triggers the `"connected"` event.
2. Spawns a dedicated goroutine for writing outgoing messages (the *Write Stream*).
3. Enters a blocking loop to read incoming messages from the socket.
4. Parses incoming JSON into `EventData` and triggers the corresponding registered callback.
5. On disconnect or error, safely shuts down the streams via the `done` channel and triggers `"disconnected"`.

### 4. Concurrency & Safety
- **Write Stream**: All outgoing messages are sent to an internal channel (`writeChannel`). A dedicated goroutine listens to this channel and writes to the TCP socket, preventing slow clients from blocking the main thread or other connections.
- **Channel Synchronization**: The library uses a `done` channel pattern and `sync.Once` to prevent `panic` scenarios (such as writing to a closed channel) during concurrent disconnects.
- **Thread-safe Mutexes**: Event storage mapping and connection states (`isConnected`) are protected against race conditions by `sync.RWMutex`.

## API Reference

### `EventData`
The standard JSON payload expected and sent by the library.
```json
{
  "eventName": "string",
  "payload": "any",
  "headers": { "key": "value" }
}
```

### `EventContext`
Passed to every event callback. Provides access to the incoming `EventData` and the `WebSocketConnection` to send responses back to the specific client.
```go
func(ctx *zap.EventContext) {
    // Access incoming payload
    data := ctx.EventData.Payload
    
    // Respond back to the client
    ctx.Connection.SendEvent(&zap.EventData{...})
}
```

### Built-in System Events
Zaptools reserves specific event names for connection lifecycle hooks:
- `"connected"`: Triggered immediately when a `ZapConnector` successfully starts.
- `"disconnected"`: Triggered when the socket closes gracefully or reading fails.
- `"error"`: Triggered on JSON parsing errors or corrupted incoming data.
- `"send-error"`: Triggered if writing data to the underlying socket fails.
