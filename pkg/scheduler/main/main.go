package main

import (
	"encoding/json"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/message"
)

func handleCreateNewPod(msg map[string]interface{}) error {
	// write pod to etcd
	// fmt.Printf("handleCreateNewPod: %s\n", msg)
	msg["node"] = "node1"
	msgs, _ := json.Marshal(msg)
	httputils.Post("http://localhost:8080/assignNodetoPod", msgs)
	return nil
}

func main() {
	go message.Consume(message.CreatePodQueueName, handleCreateNewPod)
	select {}
}
