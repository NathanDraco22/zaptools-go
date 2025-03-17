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
	contextChannel chan *EventContext 
}

func (t *EventCaller) TriggerEvent(ctx *EventContext) {
	t.contextChannel <- ctx
}

func (t *EventCaller) Run(){
	go func () {
		for ctx := range t.contextChannel {
			event, err := t.EventBook.GetEvent(ctx.EventData.EventName)
			if err != nil {
				log.Println(err)
				return
			}
			event.callback(ctx)
		}	
	}()
} 

func (t *EventCaller) Stop() {
	close(t.contextChannel)
}
//-----------------------------------

type EventProcessor struct {
	Connection *WebSocketConnection
	EventCaller *EventCaller
	StdConn StdConn
}

func (t *EventProcessor) NotifyConnected() {
	t.Connection.SetConnected(true)
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
	t.Connection.SetConnected(false)
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
	formattedPayload := fmt.Sprintf(
		"An error occurred in the event %s:\n%s", 
		originEventName, 
		err.Error(),
	)
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

func (t *EventProcessor) NotifySendError(eventSource *EventData,clientId string ,err error) {
	eventName := "send-error"
	formattedPayload := fmt.Sprintf(
		"An error occurred trying to send to the event %s to client:%s\n%s",
		eventSource.EventName,  
		clientId,
		err.Error(),
	)
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

func (t *EventProcessor) startEventStream(bufferSize int) {	
	t.EventCaller.Run()
	writeChannel := make(chan *EventData, bufferSize)
	t.Connection.writeChannel = writeChannel
	go func(eventDataChannel  <-chan *EventData ) {
		for eventData := range eventDataChannel {
			mesageData, err := json.Marshal(eventData)
			if err != nil {
				t.NotifySendError(eventData,t.Connection.Id, err)
				return
			}

			err = t.StdConn.WriteMessage(1, mesageData)
			if err != nil {
				t.Connection.SetConnected(false)
				t.NotifySendError(eventData,t.Connection.Id, err)
				return
			}
		}
	}(writeChannel)

	defer func() {
		close(writeChannel)
		t.NotifyDisconnected()
		t.EventCaller.Stop()
	}()

	t.NotifyConnected()
	
	for {
		_, data, err := t.StdConn.ReadMessage()
		if err != nil {
			t.Connection.Close()
			return
		}
		var eventData EventData
		err = json.Unmarshal(data, &eventData)
		if err != nil {
			t.NotifyError("Failed to unmarshal event data", err)
			return
		}

		if eventData.EventName == "" {
			t.NotifyError("Invalid event", fmt.Errorf("event name is empty"))
			return
		}
		ctx:= &EventContext{
			EventData: &eventData,
			Connection: t.Connection,
		}

		t.EventCaller.TriggerEvent(ctx)
	}
}