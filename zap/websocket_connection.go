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
	done         chan struct{}
	closeOnce    sync.Once
	isConnected  bool
	mu           sync.RWMutex
}

func (t *WebSocketConnection) SendEvent(eventData *EventData) error {
	if !t.GetIsConnected() {
		return errors.New("client is not connected")
	}
	select {
	case <-t.done:
		return errors.New("client is disconnected")
	case t.writeChannel <- eventData:
		return nil
	}
}

func (t *WebSocketConnection) Close() error {
	t.setConnected(false)
	t.closeOnce.Do(func() {
		close(t.done)
	})
	return t.conn.Close()
}

func (t *WebSocketConnection) GetIsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isConnected
}

func (t *WebSocketConnection) setConnected(value bool) {
	t.mu.Lock()
	t.isConnected = value
	t.mu.Unlock()
}