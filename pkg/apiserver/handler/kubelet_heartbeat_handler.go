package handler

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/protocol"
	"net/http"

	"github.com/gin-gonic/gin"
)

func KubeletHeartbeatPod(c *gin.Context) {
	var pods []protocol.Pod
	c.BindJSON(&pods)
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	for _, p := range pods {

		jsonstr, err := json.Marshal(p)
		if err != nil {
			panic(err)
		}
		st.Put(constant.EtcdPodPrefix+p.Config.Metadata.Namespace+"/"+p.Config.Metadata.Name, jsonstr)
	}

	c.JSON(http.StatusOK, nil)

}
