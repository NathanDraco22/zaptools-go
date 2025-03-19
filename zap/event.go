package zap

import (
	"fmt"
	"sync"
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
	mu sync.RWMutex
}

func (t *EventBook) SaveEvent(event Event) {
	t.events[event.Name] = event
}

func (t *EventBook) GetEvent(name string) (Event, error) {
	t.mu.RLock()
	res := t.events[name]
	t.mu.RUnlock()
	if res.callback == nil {
		return Event{}, fmt.Errorf("event %s not found", name)
	}
	return t.events[name], nil
}

//-----------------------------------

type EventRegister struct {
	eventBook *EventBook
}

func (t *EventRegister) OnEvent(name string, callback func(ctx *EventContext)) {
	event := Event{Name: name, callback: callback}
	t.eventBook.SaveEvent(event)
}

func NewEventRegister() *EventRegister {
    return &EventRegister{
        eventBook: &EventBook{
            events: make(map[string]Event),
        },
    }
}