package serverless

import (
	"encoding/json"
	"fmt"
	"io"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/protocol"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type Server struct {
	fcController *FunctionController
	route_map    map[string]string //存储function的namespace/name和service的ip的映射
	r            *gin.Engine
}

func NewServer() *Server {
	return &Server{
		fcController: NewFunctionController(),
		route_map:    make(map[string]string),
		r:            gin.Default(),
	}
}

func (s *Server) Start() {
	go s.fcController.Run()
	go s.UpdateInfo()
	s.r.POST("/triggerFunction/:functionNamespace/:functionName", s.triggerFunction)
	s.r.Run(":8050")
}

func (s *Server) UpdateInfo() {
	//每10s进行一次routine操作
	fmt.Println("Server update info")
	ticker := time.NewTicker(10 * time.Second)
	// defer ticker.Stop()
	// 开启一个goroutine执行轮询操作
	for {
		select {
		case <-ticker.C:
			s.GetServiceIpInfo()
		}
	}
}

func (s *Server) GetServiceIpInfo() {
	//从etcd中获取function的namespace/name和service的ip的映射
	fmt.Println("do GetServiceIpInfo")
	servicestr := "service"
	jsonstr, err := json.Marshal(servicestr)
	if err != nil {
		fmt.Println("json marshal failed")
		return
	}
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", jsonstr)
	var serviceList []protocol.ServiceType
	err = json.Unmarshal(resp, &serviceList)
	if err != nil {
		fmt.Println("json unmarshal failed")
		return
	}
	remoteService := make(map[string]string)
	//更新本地没有，但是etcd中有的
	for _, service := range serviceList {
		if service.Config.Metadata.Labels["type"] == "function" {
			// data, _ := yaml.Marshal(service)
			// fmt.Println("detect service is: ", string(data))

			functionName := service.Config.Metadata.Labels["functionName"]
			functionNamespace := service.Config.Metadata.Labels["functionNamespace"]
			s.route_map[functionNamespace+"/"+functionName] = service.Config.Spec.ClusterIP //是否加port？后面的TriggerFunction已经加了10000端口，这里的IP就不用再加端口了
			remoteService[functionNamespace+"/"+functionName] = service.Config.Spec.ClusterIP
		}
	}
	//删除本地有，但是etcd中没有的
	for key := range s.route_map {
		if _, ok := remoteService[key]; !ok {
			delete(s.route_map, key)
		}
	}

}

func (s *Server) triggerFunction(c *gin.Context) {
	functionName := c.Param("functionName")
	functionNamespace := c.Param("functionNamespace")
	fmt.Printf("functionName: %s, functionNamespace: %s\n", functionName, functionNamespace)
	data, _ := yaml.Marshal(s.route_map)
	fmt.Println("Now route_map is: ", string(data))

	functionServiceIP, ok := s.route_map[functionNamespace+"/"+functionName]
	if !ok {
		c.JSON(404, gin.H{"error": "Function not found"})
		return
	}
	sendPath := "http://" + functionServiceIP + ":10000"
	// resp, err := http.Post(sendPath, "application/json", c.Request.Body)
	request_body, _ := io.ReadAll(c.Request.Body)
	fmt.Printf("TriggerFunction request is: %s\n", string(request_body))
	resp := httputils.Post(sendPath, request_body)

	fmt.Println("TriggerFunction response is: ", string(resp))
	c.JSON(200, string(resp))
}
