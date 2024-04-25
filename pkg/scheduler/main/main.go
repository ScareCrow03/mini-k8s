package main

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/message"
)

func handleCreateNewPod(msg map[string]interface{}) error {
	// write pod to etcd
	fmt.Printf("handleCreateNewPod: %s\n", msg)
	res := map[string]interface{}{
		"nodeid": "node1",
	}
	msgs, _ := json.Marshal(res)
	message.Publish(message.SchedulerQueueName, msgs)
	return nil
}

func main() {
	go message.Consume(message.CreatePodQueueName, handleCreateNewPod)
	select {}
}
