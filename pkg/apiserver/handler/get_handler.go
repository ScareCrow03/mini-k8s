package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/protocol"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 因为kubelet定期向api-server更新pod状态，所以直接从etcd中取出并返回即可
// 获取所有pod并直接返回
func GetObjectByType(c *gin.Context) {
	// test_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
	var objectType string
	c.BindJSON(&objectType)

	switch objectType {
	case "Pod":
		returnValue := GetPods()
		c.JSON(http.StatusOK, returnValue)
	case "Service":
		// GetServices()
	default:
		fmt.Println("unsupported object type:", objectType)
	}

}

func GetPods() []protocol.Pod {
	fmt.Println("get pods in etcd")
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.GetWithPrefix(constant.EtcdPodPrefix)
	if err != nil {
		panic(err)
	}
	var pods []protocol.Pod
	for _, r := range reply {
		var p protocol.Pod
		err = json.Unmarshal(r.Value, &p)
		if err != nil {
			panic(err)
		}
		pods = append(pods, p)
	}
	return pods
}
