package zap

import "sync"

type StdConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type WebSocketConnection struct {
	Id           string
	conn         StdConn
	writeChannel chan<- *EventData
	IsConnected  bool
	mu           sync.Mutex
}

func (t *WebSocketConnection) SendEvent(eventData *EventData) bool {

	if !t.IsConnected {
		return false
	}
	t.writeChannel <- eventData
	return true
}

func (t *WebSocketConnection) Close() error {
	t.mu.Lock()
	t.IsConnected = false
	t.mu.Unlock()
	return t.conn.Close()
}
