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
	kubelet.Init("/home/zyc/Desktop/mini-k8s/assets/kubelet_config_worker1.yaml")
	fmt.Println(message.KubeletCreatePodQueue + "/" + kubelet.Config.Name)
	go message.Consume(message.KubeletCreatePodQueue+"/"+kubelet.Config.Name, func(msg map[string]interface{}) error {
		fmt.Println("consume: " + message.KubeletCreatePodQueue + "/" + kubelet.Config.Name)
		var pod protocol.Pod
		podjson, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("json marshal error")
			return err
		}
		json.Unmarshal(podjson, &pod.Config)
		err = kubelet2.CreatePod(&pod)
		if err != nil {
			return err
		}
		kubelet.Pods = append(kubelet.Pods, pod)
		fmt.Println("create pod finished, number of pods: ", len(kubelet.Pods))
		return nil
	})
	go message.Consume(message.KubeletDeletePodQueue+"/"+kubelet.Config.Name, func(msg map[string]interface{}) error {
		fmt.Println("consume: " + message.KubeletDeletePodQueue + "/" + kubelet.Config.Name)
		var pod protocol.Pod
		podjson, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("json marshal error")
			return err
		}
		json.Unmarshal(podjson, &pod.Config)
		err = kubelet2.DeletePod(&pod)
		if err != nil {
			return err
		}
		// remove pod from kubelet.Pods, not the same name and namespace
		j := 0
		for _, v := range kubelet.Pods {
			if v.Config.Metadata.Name != pod.Config.Metadata.Name || v.Config.Metadata.Namespace != pod.Config.Metadata.Namespace {
				kubelet.Pods[j] = v
				j++
			}
		}
		kubelet.Pods = kubelet.Pods[:j]
		fmt.Println("delete pod finished, number of pods: ", len(kubelet.Pods))
		return nil
	})
	select {}
}
