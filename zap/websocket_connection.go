package zap

type StdConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type WebSocketConnection struct {
	Id string
	conn StdConn
	writeChannel chan <- *EventData
}

func (t *WebSocketConnection) SendEvent(eventData *EventData) error {

	t.writeChannel <-eventData
	
	return nil
}


func (t *WebSocketConnection) Close() error {
	return t.conn.Close()
}



