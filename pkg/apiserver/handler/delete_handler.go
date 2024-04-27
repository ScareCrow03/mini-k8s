package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"

	"github.com/gin-gonic/gin"
)

func HandlePodStop(c *gin.Context) {
	var requestBody protocol.PodConfig
	c.BindJSON(&requestBody)
	var pod protocol.Pod
	podjson, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("json marshal error")
		return
	}
	json.Unmarshal(podjson, &pod.Config)
	msg, _ := json.Marshal(pod.Config)
	nodeid := GetPodNode(pod.Config)
	message.Publish(message.KubeletStopPodQueue+"/"+nodeid, msg)
}

func HandlePodDelete(c *gin.Context) {
	var requestBody protocol.PodConfig
	c.BindJSON(&requestBody)
	var pod protocol.Pod
	podjson, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("json marshal error")
		return
	}
	json.Unmarshal(podjson, &pod.Config)
	msg, _ := json.Marshal(pod.Config)
	nodeid := GetPodNode(pod.Config)
	message.Publish(message.KubeletDeletePodQueue+"/"+nodeid, msg)
}
