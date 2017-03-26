package iface

type Imsghandle interface {
	DeliverToMsgQueue(interface{})
	DeliverToConnectionQueue(interface{})
	DoMsg(interface{})
	DoConnection(interface{})
	AddRouter(interface{})
	GetTaskQueue() chan interface{}
	StartWorkerLoop()
}
