package main

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	kubelet2 "mini-k8s/pkg/kubelet"
	"mini-k8s/pkg/prometheus/prmts_ops"
	"mini-k8s/pkg/protocol"
	"time"

	"gopkg.in/yaml.v3"
)

// 本进程只需要在运行Prometheus的节点上运行（一般是master节点），用于自动拉取api-server配置后更新Prometheus配置文件
func main() {
	fmt.Println("mini-k8s Prometheus update process started.")
	// 每15s拉取一次当前存在的Pods和Nodes
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop() // 确保计时器停止

	done := make(chan bool)

	// 启动一个goroutine来处理ticker
	go func() {
		for {
			select {
			case <-ticker.C:
				{
					fmt.Printf("do Prometheus update\n")
					// 拉取所有pods
					req, _ := json.Marshal("pod")
					resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
					var pods []protocol.Pod
					json.Unmarshal(resp, &pods)
					// 拉取所有nodes
					req2, _ := json.Marshal("node")
					resp2 := httputils.Post(constant.HttpPreffix+"/getObjectByType", req2)
					var nodes []kubelet2.Kubelet
					json.Unmarshal(resp2, &nodes)

					cfg := prmts_ops.GetPrometheusConfigFromFile(constant.PrometheusConfigPath)
					// 从pods数组中筛选出需要暴露的Pod，并添加好endpoints映射关系
					podsJobsName2Endpoints := prmts_ops.SelectPodsNeedExposeMetrics(pods)
					// 直接配置所有Nodes的NodeExporter，位于9100端口
					nodesJobsName2Endpoints := prmts_ops.SelectNodesNeedExposeMetrics(nodes)
					// 结合上述两个map
					result := prmts_ops.MergeJobsName2Endpoints(podsJobsName2Endpoints, nodesJobsName2Endpoints)

					data, _ := yaml.Marshal(result)
					fmt.Println("Now Prometheus Listen jobs" + string(data))

					// 同步配置
					prmts_ops.SyncPrometheusConfig(&cfg, result)
					// 写回Prometheus配置文件，并热更新
					prmts_ops.ApplyPrometheusConfigToFile(cfg, constant.PrometheusConfigPath, constant.PrometheusReloadUrl)
				}
			case <-done:
				return // 退出goroutine
			}
		}
	}()

	select {}

}
