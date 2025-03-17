package zap

import (
	"errors"
	"sync"
)

type StdConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type WebSocketConnection struct {
	Id           string
	conn         StdConn
	writeChannel chan<- *EventData
	isConnected  bool
	mu           sync.RWMutex
}

func (t *WebSocketConnection) SendEvent(eventData *EventData) error {

	if !t.GetIsConnected() {
		return errors.New("not connected")
	}
	t.writeChannel <- eventData
	return nil
}

func (t *WebSocketConnection) Close() error {
	t.mu.Lock()
	t.isConnected = false
	t.mu.Unlock()
	return t.conn.Close()
}

func (t *WebSocketConnection) GetIsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isConnected
}

func (t *WebSocketConnection) SetConnected(value bool) {
	t.mu.Lock()
	t.isConnected = value
	t.mu.Unlock()
}