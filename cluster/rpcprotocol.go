package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"io"
)

type RpcServerProtocol struct {
	rpcdatapack  *RpcDataPack
	rpcmsghandle *RpcMsgHandle
}

func NewRpcServerProtocol() *RpcServerProtocol {
	return &RpcServerProtocol{
		rpcdatapack:  &RpcDataPack{},
		rpcmsghandle: NewRpcMsgHandle(),
	}
}

func (this *RpcServerProtocol) GetDataPack() iface.Idatapack {
	return this.rpcdatapack
}

func (this *RpcServerProtocol) GetMsgHandle() iface.Imsghandle {
	return this.rpcmsghandle
}

func (this *RpcServerProtocol) AddRpcRouter(router interface{}) {
	this.rpcmsghandle.AddRouter(router)
}

func (this *RpcServerProtocol) OnConnectionMade(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("node ID: %d connected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnClusterConnectioned(fconn)
}

func (this *RpcServerProtocol) OnConnectionLost(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("node ID: %d disconnected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnClusterClosed(fconn)
}

func (this *RpcServerProtocol) StartReadThread(fconn iface.Iconnection) {
	logger.Debug("start receive rpc data from socket...")
	for {
		//read per head data
		headdata := make([]byte, this.rpcdatapack.GetHeadLen())

		if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
			logger.Error(err)
			fconn.Stop()
			return
		}
		pkgHead, err := this.rpcdatapack.Unpack(headdata)
		if err != nil {
			fconn.Stop()
			return
		}
		//data
		pkg := pkgHead.(*RpcPackege)
		if pkg.Len > 0 {
			pkg.Data = make([]byte, pkg.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkg.Data); err != nil {
				fconn.Stop()
				return
			} else {
				rpcRequest := &RpcRequest{
					Fconn:   fconn,
					Rpcdata: &RpcData{},
				}

				err = json.Unmarshal(pkg.Data, rpcRequest.Rpcdata)

				if err != nil {
					logger.Error(err)
					fconn.Stop()
					return
				}

				logger.Debug(fmt.Sprintf("rpc call. data len: %d. MsgType: %d", pkg.Len, int(rpcRequest.Rpcdata.MsgType)))
				this.rpcmsghandle.DeliverToMsgQueue(rpcRequest)
			}
		}
	}
}

type RpcClientProtocol struct {
	rpcdatapack  *RpcDataPack
	rpcmsghandle *RpcMsgHandle
}

func NewRpcClientProtocol() *RpcClientProtocol {
	return &RpcClientProtocol{
		rpcdatapack:  &RpcDataPack{},
		rpcmsghandle: NewRpcMsgHandle(),
	}
}

func (this *RpcClientProtocol) GetDataPack() iface.Idatapack {
	return this.rpcdatapack
}

func (this *RpcClientProtocol) GetMsgHandle() iface.Imsghandle {
	return this.rpcmsghandle
}

func (this *RpcClientProtocol) AddRpcRouter(router interface{}) {
	this.rpcmsghandle.AddRouter(router)
}

func (this *RpcClientProtocol) OnConnectionMade(fconn iface.Iclient) {
	utils.GlobalObject.OnClusterCConnectioned(fconn)
}

func (this *RpcClientProtocol) OnConnectionLost(fconn iface.Iclient) {
	utils.GlobalObject.OnClusterCClosed(fconn)
}

func (this *RpcClientProtocol) StartReadThread(fconn iface.Iclient) {
	logger.Debug("start receive rpc data from socket...")
	for {
		//read per head data
		headdata := make([]byte, this.rpcdatapack.GetHeadLen())

		if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
			logger.Error(err)
			fconn.Stop(false)
			return
		}
		pkgHead, err := this.rpcdatapack.Unpack(headdata)
		if err != nil {
			fconn.Stop(false)
			return
		}
		//data
		pkg := pkgHead.(*RpcPackege)
		if pkg.Len > 0 {
			pkg.Data = make([]byte, pkg.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkg.Data); err != nil {
				fconn.Stop(false)
				return
			} else {
				rpcRequest := &RpcRequest{
					Fconn:   fconn,
					Rpcdata: &RpcData{},
				}
				err = json.Unmarshal(pkg.Data, rpcRequest.Rpcdata)
				if err != nil {
					logger.Error("json.Unmarshal error!!!")
					fconn.Stop(false)
					return
				}

				logger.Debug(fmt.Sprintf("rpc call. data len: %d. MsgType: %d", pkg.Len, rpcRequest.Rpcdata.MsgType))
				if rpcRequest.Rpcdata.MsgType == RESPONSE {
					//rpc result不能放主线程处理， 会超时
					this.rpcmsghandle.DoMsg(rpcRequest)
				} else {
					//请求rpc要放到主线程处理
					this.rpcmsghandle.DeliverToMsgQueue(rpcRequest)
				}
			}
		}
	}
}
