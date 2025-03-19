package zap

import (
	"encoding/json"
	"fmt"
)

type EventProcessor struct {
	Connection  *WebSocketConnection
	EventCaller *EventCaller
	StdConn     StdConn
}

func (t *EventProcessor) createPrimaryContext() *EventContext {
	eventData := &EventData{}
	ctx := &EventContext{
		EventData:  eventData,
		Connection: t.Connection,
	}
	return ctx
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
		Payload:   formattedPayload,
		Headers:   make(map[string]interface{}),
	}
	ctx := &EventContext{
		EventData:  eventData,
		Connection: t.Connection,
	}
	t.EventCaller.TriggerEventSync(ctx)
}

func (t *EventProcessor) NotifySendError(eventSource *EventData, clientId string, err error) {
	eventName := "send-error"
	formattedPayload := fmt.Sprintf(
		"An error occurred trying to send to the event %s to client:%s\n%s",
		eventSource.EventName,
		clientId,
		err.Error(),
	)
	eventData := &EventData{
		EventName: eventName,
		Payload:   formattedPayload,
		Headers:   make(map[string]interface{}),
	}
	ctx := &EventContext{
		EventData:  eventData,
		Connection: t.Connection,
	}
	t.EventCaller.TriggerEventSync(ctx)
}

func (t *EventProcessor) startEventStream(bufferSize int) {
	readyChannel := make(chan bool)
	primaryCtx := t.createPrimaryContext()
	t.EventCaller.Run(readyChannel, primaryCtx)

	writeChannel := make(chan *EventData, bufferSize)
	t.Connection.writeChannel = writeChannel

	go func(eventDataChannel <-chan *EventData) {
		for eventData := range eventDataChannel {
			mesageData, err := json.Marshal(eventData)
			if err != nil {
				t.NotifySendError(eventData, t.Connection.Id, err)
				return
			}
			err = t.StdConn.WriteMessage(1, mesageData)
			if err != nil {
				t.Connection.setConnected(false)
				t.NotifySendError(eventData, t.Connection.Id, err)
				return
			}
		}
	}(writeChannel)

	defer func() {
		t.EventCaller.Stop()
	}()

	<-readyChannel

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
		ctx := &EventContext{
			EventData:  &eventData,
			Connection: t.Connection,
		}

		t.EventCaller.TriggerEvent(ctx)
	}
}