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
	pbdatapack   *PBDataPack
	msgHandleObj *MsgHandle
}

func NewProtocol() *Protocol {
	return &Protocol{
		pbdatapack:   &PBDataPack{},
		msgHandleObj: NewMsgHandle(),
	}
}

func (this *Protocol) GetMsgHandle() iface.Imsghandle {
	return this.msgHandleObj
}

func (this *Protocol) GetDataPack() iface.Idatapack {
	return this.pbdatapack
}

func (this *Protocol) AddRpcRouter(router interface{}) {
	this.msgHandleObj.AddRouter(router)
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
		headdata := make([]byte, this.pbdatapack.GetHeadLen())

		if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
			logger.Error(err)
			fconn.Stop()
			return
		}
		pkgHead, err := this.pbdatapack.Unpack(headdata)
		if err != nil {
			fconn.Stop()
			return
		}
		//data
		pkg := pkgHead.(*PkgData)
		if pkg.Len > 0 {
			pkg.Data = make([]byte, pkg.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkg.Data); err != nil {
				fconn.Stop()
				return
			}
		}

		logger.Debug(fmt.Sprintf("msg id :%d, data len: %d", pkg.MsgId, pkg.Len))
		this.msgHandleObj.DeliverToMsgQueue(&PkgAll{
			Pdata: pkg,
			Fconn: fconn,
		})
	}
}
