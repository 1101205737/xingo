package cluster

/*
	regest rpc
*/
import (
	"fmt"
	"github.com/viphxin/xingo/fnet"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"reflect"
	"time"
)

type RpcMsgHandle struct {
	//TaskQueue chan *RpcRequest
	TaskQueue chan interface{}
	Apis      map[string]reflect.Value
}

func NewRpcMsgHandle() *RpcMsgHandle {
	return &RpcMsgHandle{
		TaskQueue: make(chan interface{}, utils.GlobalObject.MaxWorkerLen),
		Apis:      make(map[string]reflect.Value),
	}
}

func (this *RpcMsgHandle) GetTaskQueue() chan interface{} {
	return this.TaskQueue
}

/*
处理rpc消息
*/
func (this *RpcMsgHandle) DoMsg(request interface{}) {
	rpcrequest := request.(*RpcRequest)
	if rpcrequest.Rpcdata.MsgType == RESPONSE && rpcrequest.Rpcdata.Key != "" {
		//放回异步结果
		AResultGlobalObj.FillAsyncResult(rpcrequest.Rpcdata.Key, rpcrequest.Rpcdata)
		return
	} else {
		//rpc 请求
		if f, ok := this.Apis[rpcrequest.Rpcdata.Target]; ok {
			//存在
			st := time.Now()
			if rpcrequest.Rpcdata.MsgType == REQUEST_FORRESULT {
				ret := f.Call([]reflect.Value{reflect.ValueOf(request)})
				packdata, err := utils.GlobalObject.RpcSProtoc.GetDataPack().Pack(0, &RpcData{
					MsgType: RESPONSE,
					Result:  ret[0].Interface().(map[string]interface{}),
					Key:     rpcrequest.Rpcdata.Key,
				})
				if err == nil {
					rpcrequest.Fconn.Send(packdata)
				} else {
					logger.Error(err)
				}
			} else if rpcrequest.Rpcdata.MsgType == REQUEST_NORESULT {
				f.Call([]reflect.Value{reflect.ValueOf(request)})
			}

			logger.Debug(fmt.Sprintf("rpc %s cost total time: %f ms", rpcrequest.Rpcdata.Target, time.Now().Sub(st).Seconds()*1000))
		} else {
			logger.Error(fmt.Sprintf("not found rpc:  %s", rpcrequest.Rpcdata.Target))
		}
	}
}

func (this *RpcMsgHandle) DoConnection(data interface{}) {
	pkg := data.(*fnet.ConnectionQueueMsg)
	if pkg.ConnType == fnet.CONNECTIONIN {
		utils.GlobalObject.TcpServer.GetConnectionMgr().Add(pkg.Conn)
	} else {
		utils.GlobalObject.TcpServer.GetConnectionMgr().Remove(pkg.Conn)
	}
}

func (this *RpcMsgHandle) DeliverToMsgQueue(request interface{}) {
	logger.Debug("add to rpc worker.")
	this.TaskQueue <- request
}

func (this *RpcMsgHandle) DeliverToConnectionQueue(data interface{}) {
	logger.Debug("add to connection queue")
	utils.GlobalObject.TcpServer.GetConnectionQueue() <- data
}

func (this *RpcMsgHandle) AddRouter(router interface{}) {
	value := reflect.ValueOf(router)
	tp := value.Type()
	for i := 0; i < value.NumMethod(); i += 1 {
		name := tp.Method(i).Name

		if _, ok := this.Apis[name]; ok {
			//存在
			panic("repeated rpc " + name)
		}
		this.Apis[name] = value.Method(i)
		logger.Info("add rpc " + name)
	}
}

/*
开启逻辑主线程
*/
func (this *RpcMsgHandle) StartWorkerLoop() {
	logger.Info("init cluster main logic thread.")
	//根节点或者网关
	tcpserver := utils.GlobalObject.TcpServer
	protoc := utils.GlobalObject.Protoc
	if tcpserver != nil {
		for {
			select {
			case rpcData := <-this.TaskQueue:
				this.DoMsg(rpcData)
				continue
			case apiData := <-protoc.GetMsgHandle().GetTaskQueue():
				if utils.GlobalObject.RpcSProtoc == nil {
					//网关
					protoc.GetMsgHandle().DoMsg(apiData)
				} else {
					//内部节点
					this.DoMsg(apiData)
				}
				continue
			case connectionData := <-tcpserver.GetConnectionQueue():
				protoc.GetMsgHandle().DoConnection(connectionData)
				continue
			case delayTask := <-utils.GlobalObject.GsTimeScheduel.GetTriggerChannel():
				delayTask.Call()
				continue
			}
		}
	} else {
		for {
			select {
			case rpcData := <-this.TaskQueue:
				this.DoMsg(rpcData)
				continue
			case delayTask := <-utils.GlobalObject.GsTimeScheduel.GetTriggerChannel():
				delayTask.Call()
				continue
			}
		}
	}
}
