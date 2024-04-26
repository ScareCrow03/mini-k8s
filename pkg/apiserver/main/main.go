package main

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"mini-k8s/pkg/utils/uid"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/createFromFile", func(c *gin.Context) {
		var requestBody protocol.PodConfig
		c.BindJSON(&requestBody)
		fmt.Println(requestBody.Kind)
		c.JSON(http.StatusOK, gin.H{
			"message": "create resource from file:" + requestBody.Kind,
		})
		fmt.Println("write pod to etcd")
		msg, _ := json.Marshal(requestBody)
		message.Publish(message.CreatePodQueueName, msg)
	})
	r.POST("/applyFromFile", func(c *gin.Context) {
		fmt.Println((c.Request.Body))
		var requestBody map[string]interface{}
		c.BindJSON(&requestBody)
		filepath := requestBody["filepath"].(string)
		c.JSON(http.StatusOK, gin.H{
			"message": "apply resource from file:" + filepath,
		})
		fmt.Println("apply resource from file:", filepath)
		fmt.Println("update pod to etcd")
		msg, _ := json.Marshal(requestBody)
		message.Publish(message.UpdatePodQueueName, msg)
	})
	r.POST("/assignNodetoPod", func(c *gin.Context) {
		test_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
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
		if nodeid == "node1" {
			err := test_service.CreatePod(&pod)
			if err != nil {
				fmt.Printf("Failed to create pod: %v", err)
			}
		}
	})
	// go message.Consume(message.SchedulerQueueName, func(msg map[string]interface{}) error {
	// 	test_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
	// 	nodeid := msg["node"].(string)
	// 	delete(msg, "node")
	// 	var pod protocol.Pod
	// 	podjson, err := json.Marshal(msg)
	// 	if err != nil {
	// 		fmt.Println("json marshal error")
	// 		return err
	// 	}
	// 	json.Unmarshal(podjson, &pod.Config)
	// 	pod.Config.Metadata.UID = "mini-k8s_test-uid" + uid.NewUid()
	// 	if nodeid == "node1" {
	// 		err := test_service.CreatePod(&pod)
	// 		if err != nil {
	// 			fmt.Printf("Failed to create pod: %v", err)
	// 		}
	// 	}
	// 	return nil
	// })
	r.Run(":8080")
}
