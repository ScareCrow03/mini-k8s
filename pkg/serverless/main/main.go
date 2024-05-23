package main

import "mini-k8s/pkg/serverless"

func main() {
	serverless := serverless.NewServer()
	serverless.Start()
}
