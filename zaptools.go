package zaptools

import (
	zap "github.com/NathanDraco22/zaptools-go/src"
)

func NewConnector(register *zap.EventRegister, stdConn zap.StdConn) *zap.ZapConnector {
	return &zap.ZapConnector{
		Register: register,
		StdConn: stdConn,}
}

func NewRegister() *zap.EventRegister {
	return zap.NewEventRegister()
}