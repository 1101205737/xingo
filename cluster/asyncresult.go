package cluster

import (
	"fmt"
	"os"
	"time"
	"errors"
	"github.com/viphxin/xingo/logger"
	"sync"
)

type AsyncResult struct {
	key string
	result chan *RpcData
}

func GenUUID() string{
    f, _ := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
    b := make([]byte, 16)
    f.Read(b)
    f.Close()
    return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
func NewAsyncResult() *AsyncResult{
	return &AsyncResult{
		key: GenUUID(),
		result: make(chan *RpcData, 1),
	}
}

func (this *AsyncResult)GetKey() string{
	return this.key
}

func (this *AsyncResult)SetResult(data *RpcData){
	this.result <- data
}

func (this *AsyncResult)GetResult(timeout time.Duration) (*RpcData, error){
	select {
		case <-time.After(timeout):
			logger.Error(fmt.Sprintf("GetResult AsyncResult: timeout %s", this.key))
			close(this.result)
			return &RpcData{}, errors.New(fmt.Sprintf("GetResult AsyncResult: timeout %s", this.key))
		case result := <- this.result:
			return result, nil
	}
	return &RpcData{}, errors.New("GetResult AsyncResult error. reason: no")
}

var AResultGlobalObj *AsyncResultMgr = NewAsyncResultMgr()

type AsyncResultMgr struct {
	results map[string]*AsyncResult
	sync.RWMutex
}

func NewAsyncResultMgr() *AsyncResultMgr{
	return &AsyncResultMgr{
		results: make(map[string]*AsyncResult, 0),
	}
}

func (this *AsyncResultMgr)Add() *AsyncResult{
	this.Lock()
	defer this.Unlock()

	r := NewAsyncResult()
	this.results[r.GetKey()] = r
	return r
}

func (this *AsyncResultMgr)FillAsyncResult(key string, data *RpcData)error{
	this.RLock()
	defer this.RUnlock()

	r, ok := this.results[key]
	if ok{
		r.SetResult(data)
		return nil
	}else{
		return errors.New("no AsyncResult found!!!")
	}
}

func (this *AsyncResultMgr)Remove(key string){
	this.Lock()
	defer this.Unlock()

	delete(this.results, key)
}