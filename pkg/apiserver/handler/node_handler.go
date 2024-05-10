package handler

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	kubelet2 "mini-k8s/pkg/kubelet"
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
