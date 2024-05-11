package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 因为kubelet定期向api-server更新pod状态，所以直接从etcd中取出并返回即可
// 获取所有pod并直接返回
func GetObjectByType(c *gin.Context) {
	// test_service := rtm.NewRemoteRuntimeService(time.Minute)
	var objectType string
	c.BindJSON(&objectType)

	switch objectType {
	case "pod":
		c.JSON(http.StatusOK, GetAllPods())
	case "service":
		c.JSON(http.StatusOK, GetAllServices())
	case "replicaset":
		c.JSON(http.StatusOK, GetAllReplicasets())
	default:
		fmt.Println("unsupported object type:", objectType)
	}

}
