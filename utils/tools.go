package utils

import (
	"net/http"
	"runtime/debug"
	"github.com/viphxin/xingo/logger"
	"fmt"
	"time"
)

func HttpRequestWrap(uri string, targat func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request){
	return func(response http.ResponseWriter, request *http.Request) {
		defer func() {
				if err := recover(); err != nil {
					debug.PrintStack()
					logger.Info("===================http server panic recover===============")
				}
			}()
		st := time.Now()
		logger.Debug("User-Agent: ", request.Header["User-Agent"])
		targat(response, request)
		logger.Debug(fmt.Sprintf("%s cost total time: %f ms", uri, time.Now().Sub(st).Seconds()*1000))
	}
}
