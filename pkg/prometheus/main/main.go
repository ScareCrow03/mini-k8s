package main

import (
	"mini-k8s/pkg/prometheus/update_process"
)

// 本进程只需要在运行Prometheus的节点上运行（一般是master节点），用于自动拉取api-server配置后更新Prometheus配置文件
func main() {
	up := update_process.UpdateProcess{}
	up.Start()
}
