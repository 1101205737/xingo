package clusterserver

import (
	"github.com/viphxin/xingo/fnet"
	"github.com/viphxin/xingo/utils"
	"os"
	"os/signal"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/cluster"
	"sync"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/fserver"
)

type ClusterServer struct {
	Name string
	RemoteNodesMgr *cluster.ChildMgr//子节点有
	ChildsMgr *cluster.ChildMgr//root节点有
	MasterObj *fnet.TcpClient
	Cconf *cluster.ClusterConf
	modules map[string][]interface{}//所有模块统一管理
	sync.RWMutex
}

func DoCSConnectionLost(fconn iface.Iconnection) {
	logger.Error("node disconnected from " + utils.GlobalObject.Name)
	//子节点掉线
	nodename, err := fconn.GetProperty("child")
	if err == nil{
		GlobalClusterServer.RemoveChild(nodename.(string))
	}
}

func DoCCConnectionLost(fconn iface.Iclient) {
	//父节点掉线
	rname, err := fconn.GetProperty("remote")
	if err == nil{
		GlobalClusterServer.RemoveRemote(rname.(string))
		logger.Error("remote "+ rname.(string) +" disconnected from " + utils.GlobalObject.Name)
	}
}

func NewClusterServer(name, path string) *ClusterServer{
	cconf, err := cluster.NewClusterConf(path)
	if err != nil{
		panic("cluster conf error!!!")
	}
	GlobalClusterServer = &ClusterServer{
		Name: name,
		Cconf: cconf,
		RemoteNodesMgr: cluster.NewChildMgr(),
		ChildsMgr: cluster.NewChildMgr(),
		modules: make(map[string][]interface{}, 0),
	}
	utils.GlobalObject.Name = name
	utils.GlobalObject.OnClusterClosed = DoCSConnectionLost
	utils.GlobalObject.OnClusterCClosed = DoCCConnectionLost
	utils.GlobalObject.RpcCProtoc = &cluster.RpcClientProtocol{}
	if GlobalClusterServer.Cconf.Servers[name].NetPort !=0 {
		utils.GlobalObject.Protoc = &fnet.Protocol{}
	}else{
		utils.GlobalObject.Protoc = &cluster.RpcServerProtocol{}
	}
	return GlobalClusterServer
}

func (this *ClusterServer)StartClusterServer() {
	//自动发现注册modules api
	sconf, ok := this.Cconf.Servers[utils.GlobalObject.Name]
	if ok{
		modules, ok := this.modules[sconf.Module]
		if ok{
			if modules[0] != nil{
				this.AddRouter(modules[0])
			}

			if modules[1] != nil{
				this.AddRpcRouter(modules[1])
			}
		}
	}

	//http server
	//tcp server
	serverconf, ok := this.Cconf.Servers[utils.GlobalObject.Name]
	var s *fserver.Server
	if ok{
		if serverconf.NetPort > 0{
			utils.GlobalObject.TcpPort = serverconf.NetPort
			s = fserver.NewServer()
			s.Start()
		} else if serverconf.RootPort > 0{
			utils.GlobalObject.TcpPort = serverconf.RootPort
			s = fserver.NewServer()
			s.Start()
		}
	}else{
		logger.Error("server " + utils.GlobalObject.Name +"conf not found!!!")
		return
	}
	//master
	this.ConnectToMaster()
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	logger.Info("=======", sig)
	this.MasterObj.Stop()
	s.Stop()
}

func (this *ClusterServer)ConnectToMaster(){
	master := fnet.NewTcpClient(this.Cconf.Master.Host, this.Cconf.Master.RootPort, utils.GlobalObject.RpcCProtoc)
	this.MasterObj = master
	master.Start()
	//注册到master
	rpc := cluster.NewChild(utils.GlobalObject.Name, this.MasterObj)
	response, err := rpc.CallChildForResult("TakeProxy", utils.GlobalObject.Name)
	if err == nil{
		logger.Info(response.Result)
		roots, ok := response.Result["roots"]
		if ok{
			for _, root := range roots.([]interface {}){
				this.ConnectToRemote(root.(string))
			}
		}
	}else{
		logger.Error("connected to master error: ")
		logger.Error(err)
	}
}

func (this *ClusterServer)ConnectToRemote(rname string){
	rserverconf, ok := this.Cconf.Servers[rname]
	if ok{
		rserver := fnet.NewTcpClient(rserverconf.Host, rserverconf.RootPort, utils.GlobalObject.RpcCProtoc)
		this.RemoteNodesMgr.AddChild(rname, rserver)
		rserver.Start()
		rserver.SetProperty("remote", rname)
		//takeproxy
		child, err := this.RemoteNodesMgr.GetChild(rname)
		if err == nil{
			child.CallChildNotForResult("TakeProxy", utils.GlobalObject.Name)
		}
	}else{
		//未找到节点
		logger.Error("ConnectToRemote error. " + rname + " node can`t found!!!")
	}
}

func (this *ClusterServer)AddRouter(router interface{}){
	//add api ---------------start
	utils.GlobalObject.Protoc.AddRpcRouter(router)
	//add api ---------------end
}

func (this *ClusterServer)AddRpcRouter(router interface{}){
	//add api ---------------start
	utils.GlobalObject.RpcCProtoc.AddRpcRouter(router)
	//add api ---------------end
}

/*
子节点连上来回调
*/
func (this *ClusterServer)AddChild(name string, writer iface.IWriter){
	this.Lock()
	defer this.Unlock()

	this.ChildsMgr.AddChild(name, writer)
	writer.SetProperty("child", name)
}

/*
子节点断开回调
*/
func (this *ClusterServer)RemoveChild(name string){
	this.Lock()
	defer this.Unlock()

	this.ChildsMgr.RemoveChild(name)
}

func (this *ClusterServer)RemoveRemote(name string){
	this.Lock()
	defer this.Unlock()

	this.RemoteNodesMgr.RemoveChild(name)
}

/*
注册模块到分布式服务器
*/
func (this *ClusterServer)AddModule(mname string, module interface{}, rpcmodule interface{}){
	this.modules[mname] = []interface{}{module, rpcmodule}
}
