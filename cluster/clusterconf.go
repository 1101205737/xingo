package cluster

import (
	"encoding/json"
	"io/ioutil"
	"github.com/viphxin/xingo/logger"
	"errors"
)

type ClusterServerConf struct {
	Name string
	Host string
	RootPort int
	WebPort int
	NetPort int
	Remotes []string
	Module string
}

type ClusterConf struct {
	Master *ClusterServerConf
	Servers map[string]*ClusterServerConf
}

func NewClusterConf(path string) (*ClusterConf, error){
	cconf := &ClusterConf{}
	//集群服务器配置信息
	data, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Fatal(err)
		return cconf, err
	}
	err = json.Unmarshal(data, cconf)
	if err != nil {
		logger.Fatal(err)
		return cconf, err
	} else {
		logger.Info("load cluster conf successful!!!")
	}
	return cconf, nil
}

/*
获取当前节点的父节点
*/
func (this *ClusterConf)GetRemotesByName(name string) ([]string, error){
	server, ok := this.Servers[name]
	if ok{
		return server.Remotes, nil
	}else{
		return nil, errors.New("no server found!!!")
	}
}

/*
获取当前节点的子节点
*/
func (this *ClusterConf)GetChildsByName(name string) []string{
	names := make([]string, 0)
	for sername, ser := range this.Servers{
		for _, rname := range ser.Remotes{
			if rname == name{
				names = append(names, sername)
				break
			}
		}
	}
	return names
}