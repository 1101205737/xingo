package iface

import (
	"net"
)

type Iconnection interface {
	Start()
	Stop()
	GetConnection() *net.TCPConn
	GetSessionId() uint32
	Send([]byte) error
	SendBuff([]byte) error
	RemoteAddr() net.Addr
	LostConnection()
	GetSysProperty(string) (interface{}, error)
	SetSysProperty(string, interface{})
	RemoveSysProperty(string)
	GetProperty(string) (interface{}, error)
	SetProperty(string, interface{})
	RemoveProperty(string)
	GetProtoc() IServerProtocol
}
