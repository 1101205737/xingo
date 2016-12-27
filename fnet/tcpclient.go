package fnet

import (
	"net"
	"github.com/viphxin/xingo/iface"
	"fmt"
	"github.com/viphxin/xingo/logger"
	"errors"
	"sync"
)

type TcpClient struct{
	conn *net.TCPConn
	addr *net.TCPAddr
	protoc iface.IClientProtocol
	PropertyBag  map[string]interface{}
	sendtagGuard sync.RWMutex
	propertyLock sync.RWMutex
}

func NewTcpClient(ip string, port int, protoc iface.IClientProtocol) *TcpClient{
	addr := &net.TCPAddr{
		IP: net.ParseIP(ip),
		Port: port,
		Zone: "",
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err == nil{
		client := &TcpClient{
		conn: conn,
		addr: addr,
		protoc: protoc,
		PropertyBag: make(map[string]interface{}, 0),
		}
		go client.protoc.OnConnectionMade(client)
		return client
	}else{
		panic(err)
	}

}

func (this *TcpClient)Start(){
	go this.protoc.StartReadThread(this)
}

func (this *TcpClient)Stop(){
	this.protoc.OnConnectionLost(this)
}

func (this *TcpClient)Send(data []byte) error{
	this.sendtagGuard.Lock()
	defer this.sendtagGuard.Unlock()

	if _, err := this.conn.Write(data); err != nil {
		logger.Error(fmt.Sprintf("rpc client send data error.reason: %s", err))
		return err
	}
	return nil
}

func (this *TcpClient)GetConnection() *net.TCPConn{
	return this.conn
}

func (this *TcpClient) GetProperty(key string) (interface{}, error) {
	this.propertyLock.RLock()
	defer this.propertyLock.RUnlock()

	value, ok := this.PropertyBag[key]
	if ok {
		return value, nil
	} else {
		return nil, errors.New("no property in connection")
	}
}

func (this *TcpClient) SetProperty(key string, value interface{}) {
	this.propertyLock.Lock()
	defer this.propertyLock.Unlock()

	this.PropertyBag[key] = value
}

func (this *TcpClient) RemoveProperty(key string) {
	this.propertyLock.Lock()
	defer this.propertyLock.Unlock()

	delete(this.PropertyBag, key)
}
