package main

import (
	"mini-k8s/pkg/kubeproxy/dns_server"
)

func main() {
	dns_server.Init()
	select {}
}
