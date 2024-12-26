package zaptools

import (
	"github.com/NathanDraco22/zaptools-go/zap"
)

func NewConnector(register *zap.EventRegister, stdConn zap.StdConn, connectionId string) *zap.ZapConnector {
	return &zap.ZapConnector{
		Register: register,
		StdConn: stdConn,
		ConnectionId: connectionId,
	}
}

func NewRegister() *zap.EventRegister {
	return zap.NewEventRegister()
}