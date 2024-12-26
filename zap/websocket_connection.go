package zap

import (
	"encoding/json"
)

type StdConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type WebSocketConnection struct {
	Id string
	conn StdConn
}

func (t *WebSocketConnection) SendEvent(eventName string, payload interface{}, headers map[string]interface{},) error {
	if headers == nil {
		headers = make(map[string]interface{})
	}
	eventDataMap := map[string]interface{}{
		"eventName": eventName,
		"payload": payload,
		"headers": headers,
	}
	jsonEventData, err := json.Marshal(eventDataMap) 
	if err != nil {
		return err
	}
	err = t.conn.WriteMessage(1, jsonEventData)
	if err != nil {
		return err
	}
	return nil

}


func (t *WebSocketConnection) Close() error {
	return t.conn.Close()
}



