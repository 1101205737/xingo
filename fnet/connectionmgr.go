package fnet

import (
	"errors"
	"fmt"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
)

type ConnectionType int8

const (
	CONNECTIONIN  ConnectionType = 0
	CONNECTIONOUT ConnectionType = 1
)

type ConnectionQueueMsg struct {
	ConnType ConnectionType
	Conn     iface.Iconnection
}

type ConnectionMgr struct {
	connections map[uint32]iface.Iconnection
}

func (this *ConnectionMgr) Add(conn iface.Iconnection) {
	conn.GetProtoc().OnConnectionMade(conn)
	this.connections[conn.GetSessionId()] = conn
	logger.Debug(fmt.Sprintf("Total connection: %d", len(this.connections)))
}

func (this *ConnectionMgr) Remove(conn iface.Iconnection) error {
	conn.GetProtoc().OnConnectionLost(conn)
	_, ok := this.connections[conn.GetSessionId()]
	if ok {
		delete(this.connections, conn.GetSessionId())
		logger.Info(len(this.connections))
		return nil
	} else {
		return errors.New("not found!!")
	}

}

func (this *ConnectionMgr) Get(sid uint32) (iface.Iconnection, error) {
	v, ok := this.connections[sid]
	if ok {
		delete(this.connections, sid)
		return v, nil
	} else {
		return nil, errors.New("not found!!")
	}
}

func (this *ConnectionMgr) Len() int {
	return len(this.connections)
}

func NewConnectionMgr() *ConnectionMgr {
	return &ConnectionMgr{
		connections: make(map[uint32]iface.Iconnection),
	}
}
