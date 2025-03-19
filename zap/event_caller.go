package zap

import "log"

//-----------------------------------

type EventCaller struct {
	eventBook *EventBook
	contextChannel chan *EventContext
}

func (t *EventCaller) TriggerEvent(ctx *EventContext) {
	t.contextChannel <- ctx
}
func (t *EventCaller) TriggerEventSync(ctx *EventContext) {
	event, err := t.eventBook.GetEvent(ctx.EventData.EventName)
	if err != nil {
		log.Println(err)
		return
	}
    event.callback(ctx)
}
func (t *EventCaller) Run(ready chan<- bool, primaryEventCtx *EventContext) {
	t.contextChannel = make(chan *EventContext)
	go func () {
		defer func() {
			primaryEventCtx.Connection.setConnected(false)
			endEvent , err := t.eventBook.GetEvent("disconnected")
			if err == nil {
				primaryEventCtx.EventData.EventName = "disconnected"
				endEvent.callback(primaryEventCtx)
			}
		}()

		primaryEventCtx.Connection.setConnected(true)
		ready <- true
	 	initEvent, err := t.eventBook.GetEvent("connected")
		if err == nil {
			primaryEventCtx.EventData.EventName = "connected"
			initEvent.callback(primaryEventCtx)
		}

        for ctx := range t.contextChannel {
            event, err := t.eventBook.GetEvent(ctx.EventData.EventName)
            if err != nil {
                log.Println(err)
				continue
            }
            event.callback(ctx)
        }
		
    }()
}
func (t *EventCaller) Stop() {
    close(t.contextChannel)
}
//-----------------------------------
