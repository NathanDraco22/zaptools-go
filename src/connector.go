package src

import "log"


type ZapConnector struct {
	Register *EventRegister
	StdConn StdConn
	ConnectionId string
}

func (t *ZapConnector) Start() {
	log.Println("Instance Event Caller")
	eventCaller := &EventCaller{
		EventBook: t.Register.EventBook,
	}

	log.Println(t.StdConn)
	log.Println("Generate ID")
	connId := t.ConnectionId
	if connId == "" {
		connId = GenerateID()
	}

	log.Println("Event Processor")
	eventProcessor := &EventProcessor{
		Connection: &WebSocketConnection{
			Id: connId,
			conn: t.StdConn,
		},
		EventCaller: eventCaller,
		StdConn: t.StdConn,
	}
	log.Println("Start EVent Stream")
	eventProcessor.StartEventStream()
}