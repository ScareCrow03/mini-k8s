package serverless

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/type_cast"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type Server struct {
	fcController     *FunctionController
	route_map        map[string]string //存储function的namespace/name和service的ip的映射
	r                *gin.Engine
	freq             map[string]([]time.Time) // 每个function一个队列，记录最近的访问
	queuePeriod      time.Duration            // 队列中保存的时间长度
	scale            map[string]int           // 每个function的pod数量
	lastVisitTime    map[string]time.Time     // 每个function的最后一次访问
	zeroPeriod       time.Duration            // scale to zero的时间长度
	lastScaleTime    map[string]time.Time     // 最后一次scale的时间
	scalePeriod      time.Duration
	apiServerAddress string
	// workflow_map map[string]protocol.CRType // 存储workflow的namespace/name到workflow结构体的映射；它并不需要周期性获取，因为在本地没有找到时可以找api-server要一次，如果再没有找到那么返回错误
}

func NewServer() *Server {
	return &Server{
		fcController:     NewFunctionController(),
		route_map:        make(map[string]string),
		r:                gin.Default(),
		freq:             make(map[string]([]time.Time)),
		queuePeriod:      time.Minute * 1,
		scale:            make(map[string]int),
		lastVisitTime:    make(map[string]time.Time),
		zeroPeriod:       time.Minute * 2,
		lastScaleTime:    make(map[string]time.Time),
		scalePeriod:      time.Second * 15,
		apiServerAddress: "http://192.168.172.128:8080",
		// workflow_map: make(map[string]protocol.CRType),
	}
}

func (s *Server) Start() {
	fmt.Printf("Base Image is: %s\n", constant.BaseImage)
	go s.fcController.Run()
	go s.UpdateInfo()
	s.r.POST("/triggerFunction/:functionNamespace/:functionName", s.triggerFunction)
	s.r.POST("/triggerWorkflow/:workflowNamespace/:workflowName", s.TriggerWorkflow)
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
			name := functionNamespace + "/" + functionName
			s.route_map[name] = service.Config.Spec.ClusterIP //是否加port？后面的TriggerFunction已经加了10000端口，这里的IP就不用再加端口了

			remoteService[name] = service.Config.Spec.ClusterIP

			if s.freq[name] == nil {
				s.freq[name] = make([]time.Time, 0)
				s.scale[name] = 0
				s.lastVisitTime[name] = time.Now().Add(-60 * time.Minute)
				s.lastScaleTime[name] = time.Now().Add(-60 * time.Minute)
			}

		}
	}
	//删除本地有，但是etcd中没有的
	for key := range s.route_map {
		if _, ok := remoteService[key]; !ok {
			delete(s.route_map, key)
			delete(s.freq, key)
			delete(s.scale, key)
			delete(s.lastVisitTime, key)
			delete(s.lastScaleTime, key)
		}
	}

}

func (s *Server) triggerFunction(c *gin.Context) {
	functionName := c.Param("functionName")
	functionNamespace := c.Param("functionNamespace")
	fmt.Printf("functionName: %s, functionNamespace: %s\n", functionName, functionNamespace)
	data, _ := yaml.Marshal(s.route_map)
	fmt.Println("Now route_map is: ", string(data))

	name := functionNamespace + "/" + functionName
	functionServiceIP, ok := s.route_map[name]

	s.visitFunction(name)

	if !ok {
		c.JSON(404, gin.H{"error": "Function not found"})
		return
	}
	sendPath := "http://" + functionServiceIP + ":10000"
	fmt.Println("triggerFunction", sendPath)
	// resp, err := http.Post(sendPath, "application/json", c.Request.Body)
	request_body, _ := io.ReadAll(c.Request.Body)
	fmt.Printf("TriggerFunction request is: %s\n", string(request_body))
	resp := httputils.Post(sendPath, request_body)
	for {
		if resp != nil {
			break
		}
		time.Sleep(time.Microsecond * 100)
		resp = httputils.Post(sendPath, request_body)
	}
	// if string(resp) == "{\"error\":\"Function not found\"}" {

	// }

	fmt.Println("TriggerFunction response is: ", string(resp))
	c.JSON(200, string(resp))
}

func (s *Server) visitFunction(name string) {
	s.freq[name] = append(s.freq[name], time.Now())
	fmt.Println("visitFunction", name, "visitTimes:", len(s.freq[name]), "currentScale:", s.scale[name])

	s.lastVisitTime[name] = time.Now()

	scaleNew := int(math.Ceil(math.Sqrt(float64(len(s.freq[name])))))
	scaleOld := s.scale[name]
	if scaleOld == 0 || scaleNew > scaleOld { // || time.Since(s.lastScaleTime[name]) > s.scalePeriod
		s.scale[name] = scaleNew
		// TODO: scale to scaleNew
		fmt.Println(name, "scale to scaleNew", scaleNew)

		s.lastScaleTime[name] = time.Now()

		nameStr := strings.Split(name, "/")
		req, err := json.Marshal(protocol.ReplicasetSimpleType{Namespace: nameStr[0], Name: nameStr[1], Replicas: scaleNew})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		httputils.Post(s.apiServerAddress+"/changeReplicasetNum", req)

		if scaleOld == 0 {
			// 从无到有，必须保证第一个pod创建完成，kubelet将pod ip通过心跳发给api-server，proxy设置好规则，且server获取到了ip
			time.Sleep(time.Second * 5)
			httputils.Post(constant.HttpPreffix+"/serviceCheckNow", nil)
			time.Sleep(time.Second * 5)
			s.GetServiceIpInfo()
		}
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
	fmt.Println("scaleFunction", name, "visitTimes:", len(q), "currentScale:", s.scale[name])
	s.freq[name] = q

	nameStr := strings.Split(name, "/")
	if s.scale[name] > 0 && time.Since(s.lastVisitTime[name]) > s.zeroPeriod {
		// TODO: scale to zero
		fmt.Println(name, "scale to zero")

		s.lastScaleTime[name] = time.Now()

		s.scale[name] = 0
		req, err := json.Marshal(protocol.ReplicasetSimpleType{Namespace: nameStr[0], Name: nameStr[1], Replicas: s.scale[name]})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		httputils.Post(s.apiServerAddress+"/changeReplicasetNum", req)
	}
	// 不需要缩容
	// else {
	// 	scaleNew := int(max(1, math.Ceil(math.Sqrt(float64(len(q))))))
	// 	scaleOld := s.scale[name]
	// 	if scaleNew == scaleOld {
	// 		return
	// 	}
	// 	if scaleNew < scaleOld-2 || time.Since(s.lastScaleTime[name]) > s.scalePeriod { // 防止过于频繁的scale
	// 		s.scale[name] = scaleNew
	// 		// TODO: scale to scaleNew
	// 		fmt.Println(name, "scale to scaleNew", scaleNew)
	// 		req, err := json.Marshal(protocol.ReplicasetSimpleType{Namespace: nameStr[0], Name: nameStr[1], Replicas: scaleNew})
	// 		if err != nil {
	// 			fmt.Println(err.Error())
	// 			return
	// 		}
	// 		httputils.Post(s.apiServerAddress+"/changeReplicasetNum", req)
	// 	}
	// }
}

func (s *Server) checkAllFunction() {
	for key := range s.route_map {
		s.scaleFunction(key)
	}
}

// Workflow并不需要额外创建底层的抽象，给了这个配置文件，当底下的func准备就绪时，就可以直接调用它们
func (s *Server) TriggerWorkflow(c *gin.Context) {
	// 从entryNode开始，递归调用
	workflowName := c.Param("workflowName")
	workflowNamespace := c.Param("workflowNamespace")
	fmt.Printf("workflowName: %s, workflowNamespace: %s\n", workflowName, workflowNamespace)

	// 建议每次都向api-server拿，因为我们没有周期性更新逻辑！
	var workflow protocol.CRType
	var workflowAllCfg protocol.CRType
	workflowAllCfg.Kind = "workflow"
	workflowAllCfg.Metadata.Name = workflowName
	workflowAllCfg.Metadata.Namespace = workflowNamespace
	workflowAllCfg.Spec = make(map[string]interface{})
	req, _ := json.Marshal(workflowAllCfg)
	resp := httputils.Post(constant.HttpPreffix+"/getOneCR", req)
	// 进行详尽的返回体出错检查，任何一步出错了都不能往下进行
	if resp == nil {
		c.JSON(404, gin.H{"error": "Workflow not found"})
		return
	}

	var workflowInEtcd protocol.CRType
	err := json.Unmarshal(resp, &workflowInEtcd)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(404, gin.H{"error": "Workflow not found"})
		return
	}
	if workflowInEtcd.Metadata.Name == "" {
		c.JSON(404, gin.H{"error": "Workflow not found"})
		return
	}

	// 现在从etcd获取到正确的workflow对象了，放入缓存
	// s.workflow_map[workflowNamespace+"/"+workflowName] = workflowInEtcd
	// 然后再赋值给workflow对象，保证下面能正确用到
	workflow = workflowInEtcd

	// 接下来是运行状态机
	// 允许用户在触发时额外指定参数，如果没有，那么采用workflow对象中的默认参数
	// 如果用户要发，请发正确！
	var workflowSpec protocol.WorkflowSpec
	data, _ := yaml.Marshal(workflow)
	fmt.Println("Now workflow is: ", string(data))

	err = type_cast.GetObjectFromInterface(workflow.Spec, &workflowSpec)
	if err != nil {
		fmt.Printf("Parse Workflow Spec Error: %s\n", err.Error())
		c.JSON(400, gin.H{"error": "Parse Workflow Spec Error"})
		return
	}

	request_body, _ := io.ReadAll(c.Request.Body)
	var nowParamsBytes []byte
	var nowParamsMap map[string]interface{}
	// 首先尝试获取字节形式的参数，如果能获取到，那么转成map形式备用
	// 如果请求体不为空，那么直接使用请求体作为此时的参数
	if len(request_body) > 0 {
		fmt.Printf("TriggerWorkflow request is: %s\n", string(request_body))
		nowParamsBytes = request_body
	} else if len(workflowSpec.EntryParams) > 0 {
		// 请求体为空，但是workflow对象中有默认参数，那么使用默认参数
		// 这里EntryParams自身就是一个JSON形式的字符串string
		var err error
		nowParamsBytes, err = json.Marshal(workflowSpec.EntryParams)
		if err != nil {
			fmt.Println(err.Error())
			c.JSON(400, gin.H{"error": "Parse Entry Params Str To JSON bytes Error"})
		}
	} else {
		// 上述两者都为空，那么直接使用空对象
		// 请注意JSON UNMARSHAL的时候，空对象是{}，而不能是空串的bytes！
		nowParamsBytes = []byte("{}")
	}

	err = json.Unmarshal(nowParamsBytes, &nowParamsMap)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(400, gin.H{"error": "Parse Entry Params Error"})
		return
	}

	fmt.Printf("now EntryParams is: %s\n", string(nowParamsBytes))
	fmt.Printf("now EntryNode is: %s\n", workflowSpec.EntryNode)
	// 开始递归调用
	isStartFunc := true
	nowNodeName := workflowSpec.EntryNode
	lastFuncResultMap := make(map[string]interface{})
	lastFuncResultBytes := []byte("{}")
	for {
		fmt.Printf("Workflow %s/%s goto Node %s\n", workflowNamespace, workflowName, nowNodeName)
		// 从workflow对象中获取当前节点的配置
		// 如果名字为空，或者没有这个Node，那么直接返回上一次的Func计算结果
		nowNode, ok := workflowSpec.Nodes[nowNodeName]
		if nowNodeName == "" || !ok {
			fmt.Printf("Workflow %s/%s completed! final ResponseStr is %s\n", workflowNamespace, workflowName, string(lastFuncResultBytes))
			c.JSON(200, lastFuncResultMap)
			return
		}

		if !isStartFunc {
			// 如果之前调用过函数，那么更新此时可用的参数
			nowParamsBytes = lastFuncResultBytes
			nowParamsMap = lastFuncResultMap
		}
		fmt.Printf("NowParamsMap is: %v\n", nowParamsMap)

		if nowNode.Type == "func" {

			// 查看本次调用的Func的IP
			nowFuncNamespace := nowNode.FuncNodeRef.Metadata.Namespace
			nowFuncName := nowNode.FuncNodeRef.Metadata.Name

			functionServiceIP, ok := s.route_map[nowFuncNamespace+"/"+nowFuncName]
			if !ok {
				errStr := fmt.Sprintf("In Workflow %s/%s, function %s/%s not found", workflowNamespace, workflowNamespace, nowFuncNamespace, nowFuncName)
				fmt.Println(errStr)
				c.JSON(404, gin.H{"error": errStr})
				return
			}
			sendPath := "http://" + functionServiceIP + ":10000"
			nowRequestBody := nowParamsBytes

			fmt.Printf("Workflow %s/%s start do function %s/%s, request is: %s\n", workflowNamespace, workflowName, nowFuncNamespace, nowFuncName, string(nowRequestBody))

			resp := httputils.Post(sendPath, nowRequestBody)
			// 获取响应的字节形式，应该提取为map
			var nowFuncResultMap map[string]interface{}
			err := json.Unmarshal(resp, &nowFuncResultMap)
			if err != nil {
				fmt.Println(err.Error())
				errStr := fmt.Sprintf("In Workflow %s/%s, function %s/%s response parse error", workflowNamespace, workflowNamespace, nowFuncNamespace, nowFuncName)
				fmt.Println(errStr)
				c.JSON(400, gin.H{"error": errStr})
				return
			}

			// 将这次的结果存入上一次的结果中，以便下一次的计算
			lastFuncResultBytes = resp
			lastFuncResultMap = nowFuncResultMap
			isStartFunc = false
			// 更新转移节点
			nowNodeName = nowNode.FuncNodeRef.Next
			// 休息一小会，防止太快
			time.Sleep(1 * time.Second)
			fmt.Printf("Workflow %s/%s done one function %s/%s, response is: %s\n", workflowNamespace, workflowName, nowFuncNamespace, nowFuncName, string(resp))
		} else if nowNode.Type == "choice" {

			// 从上到下获取选择条件
			isSomeConditionMatched := false
			for _, condition := range nowNode.ChoiceNodeRef.Conditons {
				// 利用govaluate库进行表达式计算
				// 获取表达式
				expressionStr := condition.Expression
				expr, err := govaluate.NewEvaluableExpression(expressionStr)
				if err != nil {
					fmt.Println(err.Error())
					errStr := fmt.Sprintf("In Workflow %s/%s, choice node %s parse expression error", workflowNamespace, workflowNamespace, nowNodeName)
					fmt.Println(errStr)
					c.JSON(400, gin.H{"error": errStr})
					return
				}

				calculateRes, _ := expr.Evaluate(nowParamsMap)
				// 默认它是false
				resIsOk := false
				// 这里计算出来可能是bool，也可能是算术（我们希望它非0时即成立），以下统一转成bool；不支持其他情况，直接返回
				if boolRes, ok := calculateRes.(bool); ok {
					resIsOk = boolRes
				} else if numRes, ok := calculateRes.(float64); ok {
					if numRes != 0 {
						resIsOk = true
					} else {
						resIsOk = false
					}
				}

				if resIsOk { // 我们要求它必须是一个逻辑表达式
					fmt.Printf("Workflow %s/%s choice node %s, expression %s matched\n", workflowNamespace, workflowName, nowNodeName, expressionStr)

					nowNodeName = condition.Next
					isSomeConditionMatched = true
					break
				} else {
					continue
				}
			}

			// 如果没有找到合适的条件，那么下一步移动到""即可
			if !isSomeConditionMatched {
				fmt.Printf("Workflow %s/%s choice node %s, no expression matched, default goto end\n", workflowNamespace, workflowName, nowNodeName)
				nowNodeName = ""
			}
		} else {
			// 不支持种类的节点，直接返回
			errStr := fmt.Sprintf("In Workflow %s/%s, node %s type %s not supported", workflowNamespace, workflowNamespace, nowNodeName, nowNode.Type)
			fmt.Println(errStr)
			c.JSON(400, gin.H{"error": errStr})
			return
		}
	}
}
