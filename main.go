package main

import (
	"fmt"
	"mini-k8s/pkg/message"
)

func main() {
	// fmt.Println("Hello, mini-k8s!")
	// var dns protocol.Dns
	// dns.Spec.Host = "test.com"
	// dns.Spec.Paths = append(dns.Spec.Paths, protocol.Path{
	// 	ServiceName: "nginx",
	// 	ServiceIp:   "10.0.0.1",
	// 	Port:        80,
	// 	SubPath:     "/",
	// })
	// nginx.WriteNginxConf(dns)
	// r := gin.Default()
	// r.GET("/hello", func(c *gin.Context) {
	// 	fmt.Println("helloQ")
	// 	c.JSON(200, gin.H{
	// 		"message": "Hello, mini-k8s!",
	// 	})
	// })
	// r.Run(":8080")
	message.Consume("hello", func(msg map[string]interface{}) error {
		fmt.Println(msg)
		return nil
	})
}
