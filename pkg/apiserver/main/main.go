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

	r.POST("/createPodFromFile", handler.CreatePod)
	r.POST("/assignNodetoPod", handler.HandlePodAssignToNode)
	r.POST("/deletePodFromFile", handler.DeletePod)

	r.POST("/updateHost", handler.HandleUpdateHost)

	r.POST("/getObjectByType", handler.GetObjectByType)

	r.POST("/kubelet/register", handler.KubeletRegister)
	r.POST("/kubelet/heartbeat", handler.KubeletHeartbeat)

	r.POST("/createDnsFromFile", handler.HandleDnsCreate)
	r.POST("/deleteDnsFromFile", handler.HandleDnsDelete)

	r.POST("/createServiceFromFile", handler.CreateService)
	r.POST("/deleteServiceFromFile", handler.DeleteService)

	r.POST("/createReplicasetFromFile", handler.CreateReplicaset)
	r.POST("/deleteReplicasetFromFile", handler.DeleteReplicaset)
	r.POST("/getOneReplicaset", handler.GetOneReplicaset)

	r.POST("/createHPAFromFile", handler.CreateHPA)
	r.POST("/deleteHPAFromFile", handler.DeleteHPA)

	// 要求发过来的数据中包含Kind字段，这样可以确定存取路径
	r.POST("/createCRFromFile", handler.CreateCR)
	r.POST("/deleteCRFromFile", handler.DeleteCR)
	// 要求给定CRType，注明Kind、Namespace、Name字段
	r.POST("/getOneCR", handler.GetOneCR)

	r.POST("/createFunctionFromFile", handler.CreateFunction)
	r.POST("/deleteFunctionFromFile", handler.DeleteFunction)
	r.POST("/applyFromFile", func(c *gin.Context) {
		var requestBody map[string]interface{}
		c.BindJSON(&requestBody)
		filepath := requestBody["filepath"].(string)

		fmt.Println("apply resource from file:", filepath)
		msg, _ := json.Marshal(requestBody)
		message.Publish(message.UpdatePodQueueName, msg)

		c.JSON(http.StatusOK, "apply resource from file: "+filepath)
	})

	// 启动kubelet感知退出的协程
	go handler.CheckNodesHealthy()
	r.Run(":8080")
}
