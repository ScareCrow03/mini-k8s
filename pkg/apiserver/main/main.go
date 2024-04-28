package main

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/message"

	// rtm "mini-k8s/pkg/remoteRuntime/runtime"

	"net/http"

	// "time"

	"mini-k8s/pkg/apiserver/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/createPodFromFile", handler.HandlePodCreate)
	r.POST("/assignNodetoPod", handler.HandlePodAssignToNode)
	r.POST("/deletePodFromFile", handler.HandlePodDelete)
	r.POST("/applyFromFile", func(c *gin.Context) {
		fmt.Println((c.Request.Body))
		var requestBody map[string]interface{}
		c.BindJSON(&requestBody)
		filepath := requestBody["filepath"].(string)

		fmt.Println("apply resource from file:", filepath)
		fmt.Println("update pod to etcd")
		msg, _ := json.Marshal(requestBody)
		message.Publish(message.UpdatePodQueueName, msg)

		c.JSON(http.StatusOK, gin.H{
			"message": "apply resource from file: " + filepath,
		})
	})

	r.POST("/kubelet/register", func(c *gin.Context) {
		// TODO: register kubelet to apiserver, write into etcd
		var requestBody map[string]interface{}
		c.BindJSON(&requestBody)

		// var kubeletjson kubelet.Kubelet
		kubeletjson, _ := json.Marshal(requestBody)

		fmt.Println("register kubelet ", kubeletjson)

		c.JSON(http.StatusOK, gin.H{
			"message": "kubelet register: " + string(kubeletjson),
		})
	})

	r.POST("/getObjectByType", handler.GetObjectByType)

	r.POST("/kubelet/heartbeat/pod", handler.KubeletHeartbeatPod)

	r.Run(":8080")
}
