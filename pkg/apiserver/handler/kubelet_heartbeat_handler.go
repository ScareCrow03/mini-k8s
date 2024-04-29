package handler

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/protocol"
	"net/http"

	kubelet2 "mini-k8s/pkg/kubelet"

	"github.com/gin-gonic/gin"
)

func KubeletHeartbeat(c *gin.Context) {
	var kubelet kubelet2.Kubelet
	c.BindJSON(&kubelet)
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	// 检验etcd中的pod信息与kubelet heartbeat是否相符，此处简单删除所有该kubelet的pod
	reply, err := st.GetWithPrefix(constant.EtcdPodPrefix)
	if err != nil {
		panic(err)
	}
	for _, r := range reply {
		var p protocol.Pod
		err = json.Unmarshal(r.Value, &p)
		if err != nil {
			panic(err)
		}
		if p.Status.NodeName == kubelet.Config.Name {
			st.Del(constant.EtcdPodPrefix + p.Config.Metadata.Namespace + "/" + p.Config.Metadata.Name)
		}
	}

	for _, p := range kubelet.Pods {

		jsonstr, err := json.Marshal(p)
		if err != nil {
			panic(err)
		}
		st.Put(constant.EtcdPodPrefix+p.Config.Metadata.Namespace+"/"+p.Config.Metadata.Name, jsonstr)
	}

	c.JSON(http.StatusOK, nil)

}
