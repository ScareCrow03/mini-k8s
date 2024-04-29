package handler

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandlePodStop(c *gin.Context) {
	var requestBody protocol.PodConfig
	c.BindJSON(&requestBody)
	var pod protocol.Pod
	podjson, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(podjson, &pod.Config)
	msg, _ := json.Marshal(pod.Config)
	nodeid := GetPodNode(pod.Config)
	message.Publish(message.KubeletStopPodQueue+"/"+nodeid, msg)

	c.JSON(http.StatusOK, gin.H{
		"message": "stop pod: " + pod.Config.Metadata.Namespace + "/" + pod.Config.Metadata.Name,
	})
}

func HandlePodDelete(c *gin.Context) {
	var requestBody protocol.PodConfig
	c.BindJSON(&requestBody)
	var pod protocol.Pod
	podjson, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(podjson, &pod.Config)
	msg, _ := json.Marshal(pod.Config)
	nodeid := GetPodNode(pod.Config)
	message.Publish(message.KubeletDeletePodQueue+"/"+nodeid, msg)

	// 将删除pod写入etcd，其实不写也行，因为kubelet发心跳包含了pod信息
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	st.Del(constant.EtcdPodPrefix + pod.Config.Metadata.Namespace + "/" + pod.Config.Metadata.Name)

	c.JSON(http.StatusOK, gin.H{
		"message": "delete pod: " + pod.Config.Metadata.Namespace + "/" + pod.Config.Metadata.Name,
	})
}
