package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NathanDraco22/zaptools-go"
	"github.com/NathanDraco22/zaptools-go/zap"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)


var register = zaptools.NewRegister()


func registEvents() {

	// Triggered when a new connection is established
	register.OnEvent("connected", func(ctx *zap.EventContext) {
		fmt.Println("new connection")
		time.Sleep(5 * time.Second)
		fmt.Println("send event pre" )
		err := ctx.Connection.SendEvent(&zap.EventData{
			EventName: "hello", 
			Payload: "Hello from ON CONNECTED",
		})
		if err != nil {
			log.Println("After send event in CONNECTED")
			log.Println(err)
		}
	})

	// Triggered when a connection is closed
	register.OnEvent("disconnected", func(ctx *zap.EventContext) {
		fmt.Println("disconnected")
	})

	// Triggered when an error occurs
	register.OnEvent("error", func(ctx *zap.EventContext) {
		log.Println("an error has occured")
	})

	// Triggered when a "hello" event is received
	register.OnEvent("hello", func(ctx *zap.EventContext) {
		fmt.Println("hello event")

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