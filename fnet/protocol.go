package fnet

import (
	"errors"
	"fmt"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"io"
)

const (
	MaxPacketSize = 1024 * 1024
)

var (
	packageTooBig = errors.New("Too many data to receive!!")
)


type PkgAll struct {
	Pdata *PkgData
	Fconn iface.Iconnection
}

type Protocol struct {
}

func (this *Protocol)AddRpcRouter(router interface{}){
	MsgHandleObj.AddRouter(router)
}

func (this *Protocol) InitWorker(poolsize int32){
	MsgHandleObj.InitWorkerPool(int(poolsize))
}

func (this *Protocol) OnConnectionMade(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("client ID: %d connected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnConnectioned(fconn)
}

func (this *Protocol) OnConnectionLost(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("client ID: %d disconnected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnClosed(fconn)
}


func (this *Protocol) StartReadThread(fconn iface.Iconnection) {
	logger.Info("start receive data from socket...")
	for {
		//read per head data
		headdata := make([]byte, DefaultDataPack.GetHeadLen())

		if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
			logger.Error(err)
			fconn.Stop()
			return
		}
		pkgHead, err := DefaultDataPack.Unpack(headdata)
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
			}
		}

		logger.Debug(fmt.Sprintf("msg id :%d, data len: %d", pkgHead.MsgId, pkgHead.Len))
		if utils.GlobalObject.IsUsePool{
			MsgHandleObj.DoMsg(&PkgAll{
				Pdata: pkgHead,
				Fconn: fconn,
			})
		}else{
			MsgHandleObj.DoMsg2(&PkgAll{
				Pdata: pkgHead,
				Fconn: fconn,
			})
		}

	}
}