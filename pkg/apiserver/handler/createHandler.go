package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandlePodCreate(c *gin.Context) {
	var requestBody protocol.PodConfig
	fmt.Println("done in handler")
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
	jsonstr, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}
	st.Put(constant.EtcdPodPrefix+requestBody.Metadata.Namespace+"/"+requestBody.Metadata.Name, jsonstr)
	msg, _ := json.Marshal(requestBody)
	message.Publish(message.CreatePodQueueName, msg)
}
