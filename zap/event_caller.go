package zap

import "log"

//-----------------------------------

type EventCaller struct {
	EventBook *EventBook
	contextChannel chan *EventContext
}

func (t *EventCaller) TriggerEvent(ctx *EventContext) {
	t.contextChannel <- ctx
}
func (t *EventCaller) TriggerEventSync(ctx *EventContext) {
	event, err := t.EventBook.GetEvent(ctx.EventData.EventName)
	if err != nil {
		log.Println(err)
		return
	}
    event.callback(ctx)
}
func (t *EventCaller) Run(ready chan<- bool, primaryEventCtx *EventContext) {
	t.contextChannel = make(chan *EventContext)
	go func () {
		primaryEventCtx.Connection.setConnected(true)
		ready <- true
	 	initEvent, err := t.EventBook.GetEvent("connected")
		if err == nil {
			primaryEventCtx.EventData.EventName = "connected"
			initEvent.callback(primaryEventCtx)
		}
        for ctx := range t.contextChannel {
            event, err := t.EventBook.GetEvent(ctx.EventData.EventName)
            if err != nil {
                log.Println(err)
                return
            }
            event.callback(ctx)
        }
		primaryEventCtx.Connection.setConnected(false)
		endEvent , err := t.EventBook.GetEvent("disconnected")
		if err == nil {
			primaryEventCtx.EventData.EventName = "disconnected"
			endEvent.callback(primaryEventCtx)
		}
    }()
}
func (t *EventCaller) Stop() {
    close(t.contextChannel)
}
//-----------------------------------
