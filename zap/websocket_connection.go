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
	writeChannel chan <- *[]byte
}

func (t *WebSocketConnection) SendEvent(eventData *EventData) error {
	
	jsonEventData, err := json.Marshal(eventData) 
	
	if err != nil {
		return err
	}

	t.writeChannel <- &jsonEventData
	
	return nil

}


func (t *WebSocketConnection) Close() error {
	return t.conn.Close()
}



