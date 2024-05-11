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
	r.POST("/getNodeNames", handler.GetNodeNames)

	r.POST("/createPodFromFile", handler.HandlePodCreate)
	r.POST("/assignNodetoPod", handler.HandlePodAssignToNode)
	r.POST("/deletePodFromFile", handler.HandlePodDelete)

	r.POST("/updateHost", handler.HandleUpdateHost)

	r.POST("/getObjectByType", handler.GetObjectByType)

	r.POST("/kubelet/register", handler.KubeletRegister)
	r.POST("/kubelet/heartbeat", handler.KubeletHeartbeat)

	r.POST("/createDnsFromFile", handler.HandleDnsCreate)

	r.POST("/createServiceFromFile", handler.CreateService)
	r.POST("/deleteServiceFromFile", handler.DeleteService)

	r.POST("/createReplicasetFromFile", handler.CreateReplicaset)
	r.POST("/deleteReplicasetFromFile", handler.DeleteReplicaset)

	r.POST("/applyFromFile", func(c *gin.Context) {
		var requestBody map[string]interface{}
		c.BindJSON(&requestBody)
		filepath := requestBody["filepath"].(string)

		fmt.Println("apply resource from file:", filepath)
		msg, _ := json.Marshal(requestBody)
		message.Publish(message.UpdatePodQueueName, msg)

		c.JSON(http.StatusOK, gin.H{
			"message": "apply resource from file: " + filepath,
		})
	})

	r.Run(":8080")
}
