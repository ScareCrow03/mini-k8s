package serverless

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/protocol"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type Server struct {
	fcController  *FunctionController
	route_map     map[string]string //存储function的namespace/name和service的ip的映射
	r             *gin.Engine
	freq          map[string]([]time.Time) // 每个function一个队列，记录最近的访问
	queuePeriod   time.Duration            // 队列中保存的时间长度
	scale         map[string]int           // 每个function的pod数量
	lastVisitTime map[string]time.Time     // 每个function的最后一次访问
	zeroPeriod    time.Duration            // scale to zero的时间长度
	lastScaleTime map[string]time.Time     // 最后一次scale的时间
	scalePeriod   time.Duration
}

func NewServer() *Server {
	return &Server{
		fcController:  NewFunctionController(),
		route_map:     make(map[string]string),
		r:             gin.Default(),
		freq:          make(map[string]([]time.Time)),
		queuePeriod:   time.Minute * 1,
		lastVisitTime: make(map[string]time.Time),
		zeroPeriod:    time.Minute * 2,
		lastScaleTime: make(map[string]time.Time),
		scalePeriod:   time.Second * 15,
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
			s.checkAllFunction()
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
			s.route_map[functionName+"/"+functionNamespace] = service.Config.Spec.ClusterIP //是否加port？后面的TriggerFunction已经加了10000端口，这里的IP就不用再加端口了
			s.freq[functionName+"/"+functionNamespace] = make([]time.Time, 0)
			s.scale[functionName+"/"+functionNamespace] = 0
			s.lastVisitTime[functionName+"/"+functionNamespace] = time.Now()
			s.lastScaleTime[functionName+"/"+functionNamespace] = time.Now()
			remoteService[functionName+"/"+functionNamespace] = service.Config.Spec.ClusterIP
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

	s.visitFunction(functionNamespace + "/" + functionName)

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

func (s *Server) visitFunction(name string) {
	s.freq[name] = append(s.freq[name], time.Now())
	scaleNew := int(math.Ceil(math.Sqrt(float64(len(s.freq[name])))))
	scaleOld := s.scale[name]
	if scaleNew > scaleOld+3 || time.Since(s.lastScaleTime[name]) > s.scalePeriod { // 防止过于频繁的scale
		s.scale[name] = scaleNew
		// TODO: scale to scaleNew
	}
}

func (s *Server) scaleFunction(name string) {
	q := s.freq[name]
	for {
		if len(q) == 0 {
			break
		}
		if time.Since(q[0]) > s.queuePeriod {
			q = q[1:]
		}
	}
	fmt.Println(q)
	s.freq[name] = q

	if time.Since(s.lastVisitTime[name]) > s.zeroPeriod {
		// TODO: scale to zero
	} else {
		scaleNew := int(math.Ceil(math.Sqrt(float64(len(q)))))
		scaleOld := s.scale[name]
		if scaleNew == scaleOld {
			return
		}
		if scaleNew < scaleOld-3 || time.Since(s.lastScaleTime[name]) > s.scalePeriod { // 防止过于频繁的scale
			s.scale[name] = scaleNew
			// TODO: scale to scaleNew
		}
	}
}

func (s *Server) checkAllFunction() {
	for key := range s.route_map {
		s.scaleFunction(key)
	}
}
