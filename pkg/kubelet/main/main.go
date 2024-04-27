package main

import (
	"encoding/json"
	"fmt"
	kubelet2 "mini-k8s/pkg/kubelet"
	message "mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
)

func main() {
	var kubelet kubelet2.Kubelet
	kubelet.Init("/home/lrh/Desktop/mini-k8s/assets/kubelet_config_worker1.yaml")
	fmt.Println("link start")
	// kubelet.Start()
	fmt.Println(message.KubeletCreatePodQueue + "/" + kubelet.Config.Name)
	go message.Consume(message.KubeletCreatePodQueue+"/"+kubelet.Config.Name, func(msg map[string]interface{}) error {
		fmt.Print("consume: " + message.KubeletCreatePodQueue + "/" + kubelet.Config.Name)
		var pod protocol.Pod
		podjson, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("json marshal error")
			return err
		}
		json.Unmarshal(podjson, &pod.Config)
		err = kubelet2.CreatePod(&pod)
		fmt.Println("create pod finished")
		if err != nil {
			return err
		}
		kubelet.Pods = append(kubelet.Pods, pod)
		return nil
	})
	select {}
}
