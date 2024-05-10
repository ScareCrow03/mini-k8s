package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
)

func handleCreateNewPod(msg map[string]interface{}) error {
	var podConfig protocol.PodConfig
	podjson, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(podjson, &podConfig)

	var nodes []string
	resp := httputils.Post("http://localhost:8080/getNodeNames", nil)
	err = json.Unmarshal(resp, &nodes)
	if err != nil {
		panic(err)
	}
	if len(nodes) == 0 {
		panic("No available node!")
	}
	podConfig.NodeName = nodes[rand.Int()%len(nodes)]
	fmt.Println("All nodes: ", nodes)

	podjson, err = json.Marshal(podConfig)
	if err != nil {
		panic(err)
	}

	httputils.Post("http://localhost:8080/assignNodetoPod", podjson)
	return nil
}

func main() {
	go message.Consume(message.CreatePodQueueName, handleCreateNewPod)
	select {}
}
