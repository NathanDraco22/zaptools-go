package zap

import (
	"encoding/json"
	"fmt"
	"log"
)

//-----------------------------------
type Event struct {
	Name string
	callback func(ctx *EventContext)
}

//-----------------------------------

type EventData struct {
	EventName string `json:"eventName"`
	Payload interface{} `json:"payload"`
	Headers map[string]interface{} `json:"headers"`
}

//-----------------------------------

type EventContext struct {
	EventData *EventData
	Connection *WebSocketConnection
}

//-----------------------------------

type EventBook struct {
	events map[string]Event
}

func (t *EventBook) SaveEvent(event Event) {
	t.events[event.Name] = event
}

func (t *EventBook) GetEvent(name string) (Event, error) {
	res := t.events[name]
	if res.callback == nil {
		return Event{}, fmt.Errorf("event %s not found", name)
	}
	return t.events[name], nil
}

//-----------------------------------

type EventRegister struct {
	EventBook *EventBook
}

func (t *EventRegister) OnEvent(name string, callback func(ctx *EventContext)) {
	event := Event{Name: name, callback: callback}
	t.EventBook.SaveEvent(event)
}

func NewEventRegister() *EventRegister {
	return &EventRegister{
		EventBook: &EventBook{
			events: make(map[string]Event),
		},
	}
}

//-----------------------------------

type EventCaller struct {
	EventBook *EventBook
}

func (t *EventCaller) TriggerEvent(ctx *EventContext) {
	event, err := t.EventBook.GetEvent(ctx.EventData.EventName)
	if err != nil {
		log.Println(err)
		return
	}
	event.callback(ctx)
}

//-----------------------------------

type EventProcessor struct {
	Connection *WebSocketConnection
	EventCaller *EventCaller
	StdConn StdConn
}

func (t *EventProcessor) NotifyConnected() {
	eventName := "connected"
	eventData := &EventData{
		EventName: eventName, 
		Payload: make(map[string]interface{}), 
		Headers: make(map[string]interface{}),
	}
	ctx:= &EventContext{
		EventData: eventData,
		Connection: t.Connection,
	}
	t.EventCaller.TriggerEvent(ctx)
}

func (t *EventProcessor) NotifyDisconnected() {
	eventName := "disconnected"
	eventData := &EventData{
		EventName: eventName, 
		Payload: make(map[string]interface{}), 
		Headers: make(map[string]interface{}),
	}
	ctx:= &EventContext{
		EventData: eventData,
		Connection: t.Connection,
	}
	t.EventCaller.TriggerEvent(ctx)
}

func (t *EventProcessor) NotifyError(originEventName string, err error) {
	eventName := "error"
	formattedPayload := fmt.Sprintf("An error occurred in the event %s: %s", originEventName, err.Error())
	eventData := &EventData{
		EventName: eventName, 
		Payload: formattedPayload, 
		Headers: make(map[string]interface{}),
	}
	ctx:= &EventContext{
		EventData: eventData,
		Connection: t.Connection,
	}
	t.EventCaller.TriggerEvent(ctx)
}

func (t *EventProcessor) StartEventStream(){
	t.NotifyConnected()
	defer func() {
		t.StdConn.Close()
		t.NotifyDisconnected()
	}()
	for {
		_, data, err := t.StdConn.ReadMessage()
		if err != nil {
			t.NotifyError("Failed to read message", err)
			t.StdConn.Close()
			t.NotifyDisconnected()
			return
		}
		var eventData EventData
		err = json.Unmarshal(data, &eventData)
		if err != nil {
			t.NotifyError("Failed to unmarshal event data", err)
			t.StdConn.Close()
			t.NotifyDisconnected()
			return
		}

		if eventData.EventName == "" {
			t.StdConn.Close()
			t.NotifyError("Invalid event", fmt.Errorf("event name is empty"))
			t.NotifyDisconnected()
			return
		}
		ctx:= &EventContext{
			EventData: &eventData,
			Connection: t.Connection,
		}
		t.EventCaller.TriggerEvent(ctx)
	}
}