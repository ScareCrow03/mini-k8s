package main

import (
	"fmt"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/nginx"
)

func main() {
	fmt.Println("Hello, mini-k8s!")
	var dns protocol.Dns
	dns.Spec.Host = "test.com"
	dns.Spec.Paths = append(dns.Spec.Paths, protocol.Path{
		ServiceName: "nginx",
		ServiceIp:   "10.0.0.1",
		Port:        80,
		SubPath:     "/",
	})
	nginx.WriteNginxConf(dns)
}
