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
	go message.Consume(message.CreateServiceQueueName,
		func(msg map[string]interface{}) error {
			var svc protocol.ServiceType

			svcjson, _ := json.Marshal(msg)
			json.Unmarshal(svcjson, &svc.Config)

			ps.OnServiceAdd(&svc)
			return nil
		})

	go message.Consume(message.DeleteServiceQueueName,
		func(msg map[string]interface{}) error {
			var svc protocol.ServiceType

			svcjson, _ := json.Marshal(msg)
			json.Unmarshal(svcjson, &svc.Config)

			ps.OnServiceDelete(&svc)
			return nil
		})

	// 每10s拉取一次当前存在的Pods
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop() // 确保计时器停止

	// 用于控制退出循环的channel
	done := make(chan bool)

	// 启动一个goroutine来处理ticker
	go func() {
		for {
			select {
			case <-ticker.C:
				{
					req, _ := json.Marshal("pod")
					resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
					var pods []protocol.Pod
					json.Unmarshal(resp, &pods)
					ps.OnPodsSync2(pods)
				}
			case <-done:
				return // 退出goroutine
			}
		}
	}()

	select {}
}
