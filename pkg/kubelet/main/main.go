package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	kubelet2 "mini-k8s/pkg/kubelet"
	message "mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"os"
	"os/exec"
	"time"
)

func main() {
	var kubelet kubelet2.Kubelet
	fmt.Println(constant.WorkDir)
	kubelet.Init(constant.WorkDir + "/assets/worker-config.yaml")
	fmt.Println(message.KubeletCreatePodQueue + "/" + kubelet.Config.Name)

	// 写一个/tmp/test_html.html文件方便挂载
	os.Create("/tmp/test_html.html")
	str := "Hello from Nginx In Pod"
	os.WriteFile("/tmp/test_html.html", []byte(str), fs.FileMode(os.O_TRUNC))
	// 启动一个NodeExporter容器，如果存在则start启动
	exec.Command("docker", "run", "-d", "--name", "minik8s-node-exporter", "-p", "9100:9100", "--net=host", "--pid=host", "-v", "/:/host:ro,rslave", "quay.io/prometheus/node-exporter:v1.8.0").Run()
	exec.Command("docker", "start", "minik8s-node-exporter").Run()

	go message.Consume(message.KubeletCreatePodQueue+"/"+kubelet.Config.Name, func(msg map[string]interface{}) error {
		kubelet.Mu.Lock()

		fmt.Println("consume: " + message.KubeletCreatePodQueue + "/" + kubelet.Config.Name)
		var pod protocol.Pod
		podjson, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(podjson, &pod.Config)
		err = kubelet.CreatePod(&pod)
		if err != nil {
			panic(err)
		}
		kubelet.Pods = append(kubelet.Pods, pod)
		fmt.Println("create pod finished, number of pods: ", len(kubelet.Pods))

		// 解决延迟问题
		kubelet.PullFromApiserver()
		kubelet.SendHeartbeat()
		// time.Sleep(time.Second * 1)
		httputils.Post(constant.HttpPreffix+"/serviceCheckNow", nil)

		kubelet.Mu.Unlock()

		return nil
	})
	go message.Consume(message.KubeletDeletePodQueue+"/"+kubelet.Config.Name, func(msg map[string]interface{}) error {
		kubelet.Mu.Lock()

		fmt.Println("consume: " + message.KubeletDeletePodQueue + "/" + kubelet.Config.Name)
		var pod protocol.Pod
		podjson, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(podjson, &pod.Config)
		err = kubelet.DeletePod(&pod)
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

		// 解决延迟问题
		kubelet.PullFromApiserver()
		kubelet.SendHeartbeat()
		// time.Sleep(time.Second * 1)
		httputils.Post(constant.HttpPreffix+"/serviceCheckNow", nil)

		kubelet.Mu.Unlock()

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
				// 每隔10秒执行的函数
				kubelet.Mu.Lock()

				kubelet.PullFromApiserver()
				kubelet.SendHeartbeat()

				kubelet.Mu.Unlock()
			case <-done:
				return // 退出goroutine
			}
		}
	}()

	// 通过关闭channel来通知goroutine退出
	// close(done)

	select {}
}
