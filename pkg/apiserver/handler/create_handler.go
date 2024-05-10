package handler

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"net/http"
	"os"

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
	var requestBody map[string]interface{}
	c.BindJSON(&requestBody)
	nodeid := requestBody["node"].(string)
	delete(requestBody, "node") // 由于scheduler往请求中添加了一个键值对node:nodeName，所以需要先将其删除，再解析为podConfig
	var pod protocol.Pod
	podjson, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(podjson, &pod.Config)
	os.Create("./test_html.yml")

	pod.Config.Metadata.UID = "mini-k8s_test-uid" + uid.NewUid()
	msg, _ := json.Marshal(pod.Config)
	message.Publish(message.KubeletCreatePodQueue+"/"+nodeid, msg)

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
