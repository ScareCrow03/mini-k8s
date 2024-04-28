package handler

import (
	"encoding/json"
	"fmt"
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
	var requestBody protocol.PodConfig
	c.BindJSON(&requestBody)
	fmt.Println(requestBody.Kind)
	c.JSON(http.StatusOK, gin.H{
		"message": "create resource from file:" + requestBody.Kind,
	})
	fmt.Println("write pod to etcd")
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	jsonstr, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}
	st.Put(constant.EtcdPodPrefix+requestBody.Metadata.Namespace+"/"+requestBody.Metadata.Name, jsonstr)
	msg, _ := json.Marshal(requestBody)
	message.Publish(message.CreatePodQueueName, msg)
}

func HandlePodAssignToNode(c *gin.Context) {
	// test_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
	var requestBody map[string]interface{}
	c.BindJSON(&requestBody)
	nodeid := requestBody["node"].(string)
	delete(requestBody, "node")
	var pod protocol.Pod
	podjson, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("json marshal error")
		return
	}
	json.Unmarshal(podjson, &pod.Config)
	os.Create("./test_html.yml")
	c.JSON(http.StatusOK, gin.H{
		"message": "assign node to pod",
	})
	pod.Config.Metadata.UID = "mini-k8s_test-uid" + uid.NewUid()
	// if nodeid == "node1" {
	// 	err := test_service.CreatePod(&pod)
	// 	if err != nil {
	// 		fmt.Printf("Failed to create pod: %v", err)
	// 	}
	// }
	msg, _ := json.Marshal(pod.Config)
	message.Publish(message.KubeletCreatePodQueue+"/"+nodeid, msg)

}
