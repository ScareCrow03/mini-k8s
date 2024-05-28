package handler

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	kubelet2 "mini-k8s/pkg/kubelet"
	"mini-k8s/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetNodeNames(c *gin.Context) {
	var nodes []string
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	reply, err := st.GetWithPrefix(constant.EtcdKubeletPrefix)
	if err != nil {
		panic(err)
	}
	for _, r := range reply {
		var k kubelet2.Kubelet
		err = json.Unmarshal(r.Value, &k)
		if err != nil {
			panic(err)
		}
		nodes = append(nodes, k.Config.Name)
	}

	c.JSON(http.StatusOK, nodes)
}

// 获取etcd里存储的所有node信息，注意这些信息都是kubelet随着register/heartbeat时写入的，只有关于Node的静态信息，而没有关于Pod的动态信息
func GetAllNodes() []kubelet2.Kubelet {
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		logger.KError("etcd get error: %v", err)
		return []kubelet2.Kubelet{}
	}
	defer st.Close()
	reply, err := st.GetWithPrefix(constant.EtcdKubeletPrefix)
	if err != nil {
		logger.KError("etcd get error: %v", err)
		return []kubelet2.Kubelet{}
	}

	var nodes []kubelet2.Kubelet
	for _, r := range reply {
		var k kubelet2.Kubelet
		err = json.Unmarshal(r.Value, &k)
		if err != nil {
			return []kubelet2.Kubelet{}
		}
		nodes = append(nodes, k)
	}
	return nodes
}
