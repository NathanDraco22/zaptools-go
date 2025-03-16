package zaptools

import (
	"github.com/NathanDraco22/zaptools-go/zap"
)

type ConnectorOptions struct {
	Register *zap.EventRegister 
	StdConn zap.StdConn 
	ConnectionId string
	BufferSize int
}

func NewConnector(options ConnectorOptions) *zap.ZapConnector {
	return &zap.ZapConnector{
		Register: options.Register,
		StdConn: options.StdConn,
		ConnectionId: options.ConnectionId,
		BufferSize: options.BufferSize,
	}
}

func NewRegister() *zap.EventRegister {
	return zap.NewEventRegister()
}