package main

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/message"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/createFromFile", func(c *gin.Context) {
		var requestBody map[string]interface{}
		c.BindJSON(&requestBody)
		filepath := requestBody["filepath"].(string)
		c.JSON(http.StatusOK, gin.H{
			"message": "create resource from file:" + filepath,
		})
		fmt.Println("create resource from file:", filepath)
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
	r.Run(":8080")
}
