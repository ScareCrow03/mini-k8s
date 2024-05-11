package handler

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandlePodCreate(c *gin.Context) {
	var podConfig protocol.PodConfig
	c.BindJSON(&podConfig)

	msg, _ := json.Marshal(podConfig)
	message.Publish(message.CreatePodQueueName, msg)

	c.JSON(http.StatusOK, gin.H{
		"message": "create pod from file: " + podConfig.Metadata.Namespace + "/" + podConfig.Metadata.Name,
	})
}

func HandlePodAssignToNode(c *gin.Context) {
	var pod protocol.Pod
	c.BindJSON(&pod.Config)
	nodeName := pod.Config.NodeName

	pod.Config.Metadata.UID = "mini-k8s-pod-" + uid.NewUid()
	msg, _ := json.Marshal(pod.Config)
	message.Publish(message.KubeletCreatePodQueue+"/"+nodeName, msg)

	// 将创建pod写入etcd，其实不写也行，因为kubelet发心跳包含了pod信息
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	jsonstr, err := json.Marshal(pod)
	if err != nil {
		panic(err)
	}
	st.Put(constant.EtcdPodPrefix+pod.Config.Metadata.Namespace+"/"+pod.Config.Metadata.Name, jsonstr)

	c.JSON(http.StatusOK, gin.H{
		"message": "assign pod to node ",
	})
}
