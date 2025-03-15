package zap


type ZapConnector struct {
	Register *EventRegister
	StdConn StdConn
	ConnectionId string
	BufferSize int
}

func (t *ZapConnector) Start() {
	eventCaller := &EventCaller{
		EventBook: t.Register.EventBook,
	}
	connId := t.ConnectionId
	if connId == "" {
		connId = GenerateID()
	}
	eventProcessor := &EventProcessor{
		Connection: &WebSocketConnection{
			Id: connId,
			conn: t.StdConn,
		},
		EventCaller: eventCaller,
		StdConn: t.StdConn,
	}
	eventProcessor.startEventStream(t.BufferSize)
}