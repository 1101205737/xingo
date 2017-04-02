package fnet

import (
	"errors"
	"fmt"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"io"
	"time"
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
	//set  connection time
	fconn.SetProperty("ctime", time.Since(time.Now()))
	utils.GlobalObject.OnConnectioned(fconn)
}

func (this *Protocol) OnConnectionLost(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("client ID: %d disconnected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnClosed(fconn)
}

func (this *Protocol) SetFrequencyControl(fconn iface.Iconnection) {
	fc0, fc1 := utils.GlobalObject.GetFrequency()
	if fc1 == "h" {
		fconn.SetSysProperty("xingo_fc", 0)
		fconn.SetSysProperty("xingo_fc0", fc0)
		fconn.SetSysProperty("xingo_fc1", time.Now().UnixNano()*1e6+int64(3600*1e3))
	} else if fc1 == "m" {
		fconn.SetSysProperty("xingo_fc", 0)
		fconn.SetSysProperty("xingo_fc0", fc0)
		fconn.SetSysProperty("xingo_fc1", time.Now().UnixNano()*1e6+int64(60*1e3))
	} else if fc1 == "s" {
		fconn.SetSysProperty("xingo_fc", 0)
		fconn.SetSysProperty("xingo_fc0", fc0)
		fconn.SetSysProperty("xingo_fc1", time.Now().UnixNano()*1e6+int64(1e3))
	}
}

func (this *Protocol) DoFrequencyControl(fconn iface.Iconnection) error {
	xingo_fc1, err := fconn.GetSysProperty("xingo_fc1")
	if err != nil {
		//没有频率控制
		return nil
	} else {
		if time.Now().UnixNano()*1e6 >= xingo_fc1.(int64) {
			//init
			this.SetFrequencyControl(fconn)
		} else {
			xingo_fc, _ := fconn.GetSysProperty("xingo_fc")
			xingo_fc0, _ := fconn.GetSysProperty("xingo_fc0")
			xingo_fc_int := xingo_fc.(int) + 1
			xingo_fc0_int := xingo_fc0.(int)
			if xingo_fc_int >= xingo_fc0_int {
				//trigger
				return errors.New(fmt.Sprintf("received package exceed limit: %s", utils.GlobalObject.FrequencyControl))
			} else {
				fconn.SetSysProperty("xingo_fc", xingo_fc_int)
			}
		}
		return nil
	}
}

func (this *Protocol) StartReadThread(fconn iface.Iconnection) {
	logger.Info("start receive data from socket...")
	//加频率控制
	this.SetFrequencyControl(fconn)
	for {
		//频率控制
		err := this.DoFrequencyControl(fconn)
		if err != nil {
			logger.Error(err)
			fconn.Stop()
			return
		}
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
