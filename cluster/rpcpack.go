package cluster

import (
	"encoding/binary"
	"fmt"
	"bytes"
	"github.com/viphxin/xingo/fnet"
	"errors"
	"encoding/json"
	"github.com/viphxin/xingo/iface"
)

type RpcData struct {
	MsgType RpcSignal `json:"msgtype"`
	Key string `json:"key,omitempty"`
	Target string `json:"target,omitempty"`
	Args []interface{} `json:"args,omitempty"`
	Result map[string]interface{} `json:"result,omitempty"`
}


type RpcPackege struct {
	Len int32
	Data []byte
}

type RpcRequest struct {
	Fconn iface.IWriter
	Rpcdata *RpcData
}

type RpcDataPack struct {

}

var DefaultRpcDataPack *RpcDataPack = &RpcDataPack{}

func (this *RpcDataPack)GetHeadLen() int32{
	return 4
}

func (this *RpcDataPack)Unpack(headdata []byte) (rp *RpcPackege, err error) {
	headbuf := bytes.NewReader(headdata)

	rp = &RpcPackege{}

	// 读取Len
	if err = binary.Read(headbuf, binary.LittleEndian, &rp.Len); err != nil {
		return nil, err
	}

	// 封包太大
	if rp.Len > fnet.MaxPacketSize {
		return nil, errors.New("rpc packege too big!!!")
	}

	return rp, nil
}

func (this *RpcDataPack)Pack(data *RpcData) (out []byte, err error) {
	outbuff := bytes.NewBuffer([]byte{})
	// 进行编码
	dataBytes := []byte{}
	if data != nil {
		dataBytes, err = json.Marshal(data)
	}

	if err != nil {
		fmt.Println(fmt.Sprintf("json marshaling error:  %s", err))
	}
	// 写Len
	if err = binary.Write(outbuff, binary.LittleEndian, uint32(len(dataBytes))); err != nil {
		return
	}


	//all pkg data
	if err = binary.Write(outbuff, binary.LittleEndian, dataBytes); err != nil {
		return
	}

	out = outbuff.Bytes()
	return

}
