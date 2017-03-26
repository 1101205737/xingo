package iface

type IClientProtocol interface {
	OnConnectionMade(fconn Iclient)
	OnConnectionLost(fconn Iclient)
	StartReadThread(fconn Iclient)
	AddRpcRouter(interface{})
	GetMsgHandle() Imsghandle
	GetDataPack() Idatapack
}

type IServerProtocol interface {
	OnConnectionMade(fconn Iconnection)
	OnConnectionLost(fconn Iconnection)
	StartReadThread(fconn Iconnection)
	AddRpcRouter(interface{})
	GetMsgHandle() Imsghandle
	GetDataPack() Idatapack
}
