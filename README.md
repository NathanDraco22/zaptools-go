<h1 align="center">Zaptools</h1>

<p align="center">
  <img src="https://raw.githubusercontent.com/NathanDraco22/zaptools-dart/main/assets/zaptools-logo-150.png" />
  <h3 align="center">
    A toolkit for Event-Driven websocket management
  <h3>
</p>

### Also Supported
| Lang               |Side  |View Source                                                                                           |
|:------------------:|:----:|:------------------------------------------------------------------------------------------------------|
|<a href="https://dart.dev" target="_blank"> <img src="https://www.vectorlogo.zone/logos/dartlang/dartlang-icon.svg" alt="python" width="25" height="25"/> </a>| Client/Server |[`zaptools_dart`](https://github.com/NathanDraco22/zaptools-dart)|
|<a href="https://www.python.org" target="_blank"> <img src="https://upload.wikimedia.org/wikipedia/commons/archive/c/c3/20220730085403%21Python-logo-notext.svg" alt="python" width="25" height="25"/> </a>| Client/Server |[`zaptools_python`](https://github.com/NathanDraco22/zaptools-python)|

### Getting Started

Zaptools provides tools for building event-driven websocket integration, compatible with gorilla/websocket and Fiber web framework.


#### installation
``` bash
go get github.com/NathanDraco22/zaptools-go
```

#### Gorilla Mux
This example is using echo framework, but you can use any framework compatible with gorilla websocket (Gin, Chi, etc)
```go
package main

import (
	"log"
	"net/http"

	"github.com/NathanDraco22/zaptools-go"
	"github.com/NathanDraco22/zaptools-go/zap"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)


var register = zaptools.NewRegister()


func registEvents() {

	// Triggered when a new connection is established
	register.OnEvent("connected", func(ctx *zap.EventContext) {
		log.Println("new connection")
	})

	// Triggered when a connection is closed
	register.OnEvent("disconnected", func(ctx *zap.EventContext) {
		log.Println("disconnected")
	})

	// Triggered when an error occurs
	register.OnEvent("error", func(ctx *zap.EventContext) {
		log.Println("an error has occured")
	})

	// Triggered when a "hello" event is received
	register.OnEvent("hello", func(ctx *zap.EventContext) {

		// Send "hello" event
		eventData := zap.EventData{
			EventName: "hello", 
			Payload: "Hello from server", 
		}
		ctx.Connection.SendEvent(&eventData)
	})
}



func main() {

	router := echo.New()

	// Register events
	registEvents()
	
	router.GET("/ws", func(c echo.Context) error {

		// Create upgrader
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

		// upgrade connection to websocket
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)

		if err != nil {
			log.Println(err)
			return err
		}
		
		// Create connector with WebSocket connection
		options := zaptools.ConnectorOptions{
			Register: register,
			StdConn: conn,
		}
		zapConnector := zaptools.NewConnector(options)
		
		// Start connector
		zapConnector.Start()

		return nil
	})

	err := router.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
```

- Firstly create a new `*zap.EventRegister`, all events triggers are registered by `*zap.EventRegister`.
- Create a websocket connection using the upgrader from gorilla websocket.
- Create a `connector.Options` and then the constructor `zaptools.NewConnector()`.
> Zaptools generate an ID if you set the ID as empty string in the `connector.Options`

### Fiber example
```go
package main

import (
	"log"

	"github.com/NathanDraco22/zaptools-go"
	"github.com/NathanDraco22/zaptools-go/zap"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)


var register = zaptools.NewRegister()


func registEvents() {

	// Triggered when a new connection is established
	register.OnEvent("connected", func(ctx *zap.EventContext) {
		log.Println("new connection")
	})

	// Triggered when a connection is closed
	register.OnEvent("disconnected", func(ctx *zap.EventContext) {
		log.Println("disconnected")
	})

	// Triggered when an error occurs
	register.OnEvent("error", func(ctx *zap.EventContext) {
		log.Println("an error has occured")
	})

	// Triggered when a "hello" event is received
	register.OnEvent("hello", func(ctx *zap.EventContext) {

		// Send "hello" event
		eventData := zap.EventData{
			EventName: "hello", 
			Payload: "Hello from server", 
		}
		ctx.Connection.SendEvent(&eventData)
		
	})
}



func main() {

	app := fiber.New()

	registEvents()

	app.Use("/ws", func(ctx *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(ctx) {
			ctx.Locals("allowed", true)
			return ctx.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(conn *websocket.Conn) {
		// Create connector with WebSocket connection
		options := zaptools.ConnectorOptions{
			Register: register,
			StdConn: conn,
		}
		zapConnector := zaptools.NewConnector(options)
		zapConnector.Start()
	}))

	log.Fatal(app.Listen(":8080"))
}
```

### StdConn Interface

Actually, all structs that implement the StdConn interface can be used with the Zaptools Connector.

```go
type StdConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}
```

### EventContext object
Each element is triggered with a `EventContext` object. This `EventContext` object contains information about the current event and which `WebSocketConnection` is invoking it.
```go
register.OnEvent("hello", func(ctx *zap.EventContext) {

		eventData := ctx.EventData
		
		eventData.EventName // name of the event
		eventData.Payload // payload of the events
		eventData.Headers // headers of the events

		// Send a response to the client
		eventData := zap.EventData{
			EventName: "hello", 
			Payload: "Hello from server", 
		}
		// The event is sending in a separated GoRoutine
		ctx.Connection.SendEvent(&eventData)
		
	})
```

### Events

The `"connected"`, `"disconnected"` and `"error"` events can be used to trigger an action when a connection is started and after it is closed or when a error ocurred in a event.

```go

var register = zaptools.NewRegister()


func registEvents() {

	// Triggered when a new connection is established
	register.OnEvent("connected", func(ctx *zap.EventContext) {
		log.Println("new connection")

	})

	// Triggered when a connection is closed
	register.OnEvent("disconnected", func(ctx *zap.EventContext) {
		log.Println("disconnected")
	})

	// Triggered when an error occurs
	register.OnEvent("error", func(ctx *zap.EventContext) {
		log.Println("an error has occured")
	})
}
```
> Error details in `payload`,  all event callbacks will be triggered in a separate goroutine

### Concurrency
When you create a new `*zap.ZapConnector` and execute `Start()` method, a dedicated GoRoutine is created to Write/Send the event to the websocket connection.
When the `SendEvent` is called, the `EventData` is tranfered to the GoRoutine dedicated to Write/Send the data to the websocker connection.
The `*zap.EventRegister` object triggers all events within the same GoRoutine as the originating HTTP request.
```Go
	// Inside GoRoutine #1
	register.OnEvent("hello", func(ctx *zap.EventContext) {

		eventData := zap.EventData{
			EventName: "hello", 
			Payload: "Hello from server", 
		}
		// will send using GoRoutine #2
		ctx.Connection.SendEvent(&eventData)
		
	})

	// GoRoutine #1
	router.GET("/ws", func(c echo.Context) error {
		///...... CODE HERE
		zapConnector := zaptools.NewConnector(options)
		zapConnector.Start()

	})
```

## Contributions are wellcome
