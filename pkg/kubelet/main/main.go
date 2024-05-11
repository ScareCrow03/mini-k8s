package main

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	kubelet2 "mini-k8s/pkg/kubelet"
	message "mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"time"
)

func main() {
	var kubelet kubelet2.Kubelet
	fmt.Println(constant.WorkDir)
	kubelet.Init(constant.WorkDir + "/assets/worker-config.yaml")
	fmt.Println(message.KubeletCreatePodQueue + "/" + kubelet.Config.Name)
	go message.Consume(message.KubeletCreatePodQueue+"/"+kubelet.Config.Name, func(msg map[string]interface{}) error {
		fmt.Println("consume: " + message.KubeletCreatePodQueue + "/" + kubelet.Config.Name)
		var pod protocol.Pod
		podjson, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(podjson, &pod.Config)
		err = kubelet2.CreatePod(&pod)
		if err != nil {
			panic(err)
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
			panic(err)
		}
		json.Unmarshal(podjson, &pod.Config)
		err = kubelet2.DeletePod(&pod)
		if err != nil {
			panic(err)
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

	// 定义一个每隔10秒触发一次的计时器
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop() // 确保计时器停止

	// 用于控制退出循环的channel
	done := make(chan bool)

	// 启动一个goroutine来处理ticker
	go func() {
		for {
			select {
			case <-ticker.C:
				// 每隔1秒执行的函数
				kubelet.SendHeartbeat()
			case <-done:
				return // 退出goroutine
			}
		}
	}()

	// 通过关闭channel来通知goroutine退出
	// close(done)

	select {}
}
