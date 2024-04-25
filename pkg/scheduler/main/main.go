package main

import (
	"fmt"
	"mini-k8s/pkg/message"
)

func handleCreateNewPod(msg map[string]interface{}) error {
	// write pod to etcd
	fmt.Printf("handleCreateNewPod: %s\n", msg)
	return nil
}

func main() {
	go message.Consume(message.CreatePodQueueName, handleCreateNewPod)
	select {}
}
