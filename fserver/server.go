package fserver

import (
	"github.com/viphxin/xingo/fnet"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/timer"
	"github.com/viphxin/xingo/utils"
	"net"
	"os"
	"os/signal"
	"time"
)

func init() {
	utils.GlobalObject.Protoc = fnet.NewProtocol()
	// --------------------------------------------init log start
	utils.ReSettingLog()
	// --------------------------------------------init log end
}

type Server struct {
	Port            int
	MaxConn         int
	GenNum          uint32
	connectionMgr   iface.Iconnectionmgr
	connectionQueue chan interface{}
}

func NewServer() iface.Iserver {
	s := &Server{
		Port:            utils.GlobalObject.TcpPort,
		MaxConn:         utils.GlobalObject.MaxConn,
		connectionMgr:   fnet.NewConnectionMgr(),
		connectionQueue: make(chan interface{}, utils.GlobalObject.MaxConnWorkerLen),
	}
	utils.GlobalObject.TcpServer = s
	return s
}

func (this *Server) handleConnection(conn *net.TCPConn) {
	this.GenNum += 1
	conn.SetNoDelay(true)
	conn.SetKeepAlive(true)
	// conn.SetDeadline(time.Now().Add(time.Minute * 2))
	//fconn := fnet.NewConnection(conn, this.GenNum, &fnet.Protocol{})
	fconn := fnet.NewConnection(conn, this.GenNum, utils.GlobalObject.Protoc)
	//fconn := fnet.NewConnection(conn, this.GenNum, &cluster.RpcServerProtocol{})
	fconn.Start()
}

func (this *Server) Start() {
	go func() {
		ln, err := net.ListenTCP("tcp", &net.TCPAddr{
			Port: this.Port,
		})
		if err != nil {
			logger.Error(err)
		}
		logger.Info("start xingo server...")
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				logger.Error(err)
			}
			//max client exceed
			if this.GetConnectionMgr().Len() >= utils.GlobalObject.MaxConn {
				conn.Close()
			} else {
				go this.handleConnection(conn)
			}
		}
	}()
}

func (this *Server) GetConnectionMgr() iface.Iconnectionmgr {
	return this.connectionMgr
}

func (this *Server) GetConnectionQueue() chan interface{} {
	return this.connectionQueue
}

func (this *Server) Stop() {
	logger.Info("stop xingo server!!!")
	utils.GlobalObject.OnServerStop()
}

func (this *Server) AddRouter(router interface{}) {
	logger.Info("AddRouter")
	utils.GlobalObject.Protoc.GetMsgHandle().AddRouter(router)
}

func (this *Server) CallLater(durations time.Duration, f func(v ...interface{}), args ...interface{}) {
	delayTask := timer.NewTimer(durations, f, args)
	delayTask.Run()
}

func (this *Server) CallWhen(ts string, f func(v ...interface{}), args ...interface{}) {
	loc, err_loc := time.LoadLocation("Local")
	if err_loc != nil {
		logger.Error(err_loc)
		return
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", ts, loc)
	now := time.Now()
	//logger.Info(t)
	//logger.Info(now)
	//logger.Info(now.Before(t) == true)
	if err == nil {
		if now.Before(t) {
			this.CallLater(t.Sub(now), f, args...)
		} else {
			logger.Error("CallWhen time before now")
		}
	} else {
		logger.Error(err)
	}
}

func (this *Server) CallLoop(durations time.Duration, f func(v ...interface{}), args ...interface{}) {
	go func() {
		delayTask := timer.NewTimer(durations, f, args)
		for {
			time.Sleep(delayTask.GetDurations())
			delayTask.GetFunc().Call()
		}
	}()
}

func (this *Server) WaitSignal() {
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	logger.Info("=======", sig)
	this.Stop()
}

func (this *Server) Serve() {
	this.Start()
	go utils.GlobalObject.Protoc.GetMsgHandle().StartWorkerLoop()
	this.WaitSignal()
}
