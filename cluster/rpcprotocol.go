package cluster

import (
	"fmt"
	"github.com/viphxin/xingo/utils"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/iface"
	"io"
	"encoding/json"
)

type RpcServerProtocol struct {
}

func NewRpcServerProtocol() *RpcServerProtocol{
	return &RpcServerProtocol{}
}

func (this *RpcServerProtocol)AddRpcRouter(router interface{}){
	RpcHandleObj.AddRouter(router)
}

func (this *RpcServerProtocol) InitWorker(poolsize int32){
	RpcHandleObj.InitWorkerPool(int(poolsize))
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
	logger.Info("start receive rpc data from socket...")
	for {
		//read per head data
		headdata := make([]byte, DefaultRpcDataPack.GetHeadLen())

		if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
			logger.Error(err)
			fconn.Stop()
			return
		}
		pkgHead, err := DefaultRpcDataPack.Unpack(headdata)
		if err != nil {
			fconn.Stop()
			return
		}
		//data
		if pkgHead.Len > 0 {
			pkgHead.Data = make([]byte, pkgHead.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkgHead.Data); err != nil {
				fconn.Stop()
				return
			}else{
				rpcRequest := &RpcRequest{
					Fconn: fconn,
					Rpcdata: &RpcData{},
				}

				err = json.Unmarshal(pkgHead.Data, rpcRequest.Rpcdata)

				if err != nil{
					logger.Error(err)
					fconn.Stop()
					return
				}

				logger.Debug(fmt.Sprintf("rpc call. data len: %d. MsgType: %d", pkgHead.Len, int(rpcRequest.Rpcdata.MsgType)))
				if utils.GlobalObject.IsUsePool{
					RpcHandleObj.DoMsg2(rpcRequest)
				}else{
					RpcHandleObj.DoMsg2(rpcRequest)
				}
			}
		}
	}
}

type RpcClientProtocol struct {
}

func (this *RpcClientProtocol)AddRpcRouter(router interface{}){
	RpcHandleObj.AddRouter(router)
}

func (this *RpcClientProtocol)OnConnectionMade(fconn iface.Iclient){
	utils.GlobalObject.OnClusterCConnectioned(fconn)
}

func (this *RpcClientProtocol)OnConnectionLost(fconn iface.Iclient){
	utils.GlobalObject.OnClusterCClosed(fconn)
}

func (this *RpcClientProtocol)StartReadThread(fconn iface.Iclient){
	logger.Info("start receive rpc data from socket...")
	for {
		//read per head data
		headdata := make([]byte, DefaultRpcDataPack.GetHeadLen())

		if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
			logger.Error(err)
			fconn.Stop()
			return
		}
		pkgHead, err := DefaultRpcDataPack.Unpack(headdata)
		if err != nil {
			fconn.Stop()
			return
		}
		//data
		if pkgHead.Len > 0 {
			pkgHead.Data = make([]byte, pkgHead.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkgHead.Data); err != nil {
				fconn.Stop()
				return
			}else{
				rpcRequest := &RpcRequest{
					Fconn: fconn,
					Rpcdata: &RpcData{},
				}
				err = json.Unmarshal(pkgHead.Data, rpcRequest.Rpcdata)
				if err != nil{
					logger.Error("json.Unmarshal error!!!")
					fconn.Stop()
					return
				}

				logger.Debug(fmt.Sprintf("rpc call. data len: %d. MsgType: %d", pkgHead.Len, rpcRequest.Rpcdata.MsgType))
				if utils.GlobalObject.IsUsePool{
					RpcHandleObj.DoMsg2(rpcRequest)
				}else{
					RpcHandleObj.DoMsg2(rpcRequest)
				}
			}
		}
	}
}