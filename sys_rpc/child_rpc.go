package sys_rpc

import (
	"github.com/viphxin/xingo/cluster"
	"github.com/viphxin/xingo/clusterserver"
	"github.com/viphxin/xingo/logger"
)

type ChildRpc struct {

}

/*
master 通知父节点上线, 收到通知的子节点需要链接对应父节点
*/
func (this *ChildRpc)RootTakeProxy(request *cluster.RpcRequest){
	logger.Info("dasdsadsadsa RootTakeProxy")
	rname := request.Rpcdata.Args[0].(string)
	clusterserver.GlobalClusterServer.ConnectToRemote(rname)
}
