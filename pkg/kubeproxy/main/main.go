package main

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/kubeproxy/proxy_server"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"time"
)

func main() {
	ps := proxy_server.NewProxyServer(constant.CLUSTER_CIDR_DEFAULT)
	ps.IpvsOps.Clear()
	go message.Consume(message.CreateServiceQueueName,
		func(msg map[string]interface{}) error {
			ps.Mu.Lock()
			var svc protocol.ServiceType

			svcjson, _ := json.Marshal(msg)
			json.Unmarshal(svcjson, &svc)

			ps.OnServiceAdd(&svc)
			ps.Mu.Unlock()
			return nil
		})

	go message.Consume(message.DeleteServiceQueueName,
		func(msg map[string]interface{}) error {
			ps.Mu.Lock()
			var svc protocol.ServiceType

			svcjson, _ := json.Marshal(msg)
			json.Unmarshal(svcjson, &svc)

			ps.OnServiceDelete(&svc)
			ps.Mu.Unlock()
			return nil
		})

	go message.Consume(message.ServiceCheckNowQueueName, // 立即更新
		func(msg map[string]interface{}) error {
			ps.Mu.Lock()
			req, _ := json.Marshal("pod")
			resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
			var pods []protocol.Pod
			json.Unmarshal(resp, &pods)

			req2, _ := json.Marshal("service")
			resp2 := httputils.Post(constant.HttpPreffix+"/getObjectByType", req2)
			var services []protocol.ServiceType
			json.Unmarshal(resp2, &services)

			ps.OnPodsAndServiceSync(pods, services)
			ps.Mu.Unlock()
			return nil
		})

	// 每10s拉取一次当前存在的Pods
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop() // 确保计时器停止

	//
	done := make(chan bool)

	// 启动一个goroutine来处理ticker
	go func() {
		for {
			select {
			case <-ticker.C:
				{
					ps.Mu.Lock()
					req, _ := json.Marshal("pod")
					resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
					var pods []protocol.Pod
					json.Unmarshal(resp, &pods)

					req2, _ := json.Marshal("service")
					resp2 := httputils.Post(constant.HttpPreffix+"/getObjectByType", req2)
					var services []protocol.ServiceType
					json.Unmarshal(resp2, &services)

					ps.OnPodsAndServiceSync(pods, services)
					ps.Mu.Unlock()
				}
			case <-done:
				return // 退出goroutine
			}
		}
	}()

	select {}
}
