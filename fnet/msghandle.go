package fnet

/*
	find msg api
*/
import (
	"fmt"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type MsgHandle struct {
	//TaskQueue chan *PkgAll
	TaskQueue chan interface{}
	Apis      map[uint32]reflect.Value
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		TaskQueue: make(chan interface{}, utils.GlobalObject.MaxWorkerLen),
		Apis:      make(map[uint32]reflect.Value),
	}
}

func (this *MsgHandle) GetTaskQueue() chan interface{} {
	return this.TaskQueue
}

func (this *MsgHandle) DeliverToMsgQueue(data interface{}) {
	logger.Debug("add to msg queue")
	this.TaskQueue <- data.(*PkgAll)
}

func (this *MsgHandle) DeliverToConnectionQueue(data interface{}) {
	logger.Debug("add to connection queue")
	utils.GlobalObject.TcpServer.GetConnectionQueue() <- data
}

func (this *MsgHandle) DoMsg(data interface{}) {
	pkg := data.(*PkgAll)
	if f, ok := this.Apis[pkg.Pdata.MsgId]; ok {
		//存在
		st := time.Now()
		//f.Call([]reflect.Value{reflect.ValueOf(data)})
		utils.XingoTry(f, []reflect.Value{reflect.ValueOf(data)}, nil)
		logger.Debug(fmt.Sprintf("Api_%d cost total time: %f ms", pkg.Pdata.MsgId, time.Now().Sub(st).Seconds()*1000))
	} else {
		logger.Error(fmt.Sprintf("not found api:  %d", pkg.Pdata.MsgId))
	}
}

func (this *MsgHandle) DoConnection(data interface{}) {
	pkg := data.(*ConnectionQueueMsg)
	if pkg.ConnType == CONNECTIONIN {
		utils.GlobalObject.TcpServer.GetConnectionMgr().Add(pkg.Conn)
	} else {
		utils.GlobalObject.TcpServer.GetConnectionMgr().Remove(pkg.Conn)
	}
}

func (this *MsgHandle) AddRouter(router interface{}) {
	value := reflect.ValueOf(router)
	tp := value.Type()
	for i := 0; i < value.NumMethod(); i += 1 {
		name := tp.Method(i).Name
		k := strings.Split(name, "_")
		index, err := strconv.Atoi(k[1])
		if err != nil {
			panic("error api: " + name)
		}
		if _, ok := this.Apis[uint32(index)]; ok {
			//存在
			panic("repeated api " + string(index))
		}
		this.Apis[uint32(index)] = value.Method(i)
		logger.Info("add api " + name)
	}

	//exec test
	// for i := 0; i < 100; i += 1 {
	// 	Apis[1].Call([]reflect.Value{reflect.ValueOf("huangxin"), reflect.ValueOf(router)})
	// 	Apis[2].Call([]reflect.Value{})
	// }
	// fmt.Println(this.Apis)
	// this.Apis[2].Call([]reflect.Value{reflect.ValueOf(&PkgAll{})})
}

/*
开启逻辑主线程
*/
func (this *MsgHandle) StartWorkerLoop() {
	logger.Info("init main logic thread.")
	for {
		select {
		case apiData := <-this.TaskQueue:
			this.DoMsg(apiData)
			continue
		case connectionData := <-utils.GlobalObject.TcpServer.GetConnectionQueue():
			this.DoConnection(connectionData)
			continue
		case delayTask := <-utils.GlobalObject.GsTimeScheduel.GetTriggerChannel():
			delayTask.Call()
			continue
		}
	}
}
