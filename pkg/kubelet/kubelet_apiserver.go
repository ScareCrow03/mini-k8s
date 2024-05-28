package kubelet

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"time"
)

func CheckPodSpec(p protocol.PodConfig) bool {
	if p.Metadata.Name == "" || p.Metadata.Namespace == "" {
		return false
	}
	for _, c := range p.Spec.Containers {
		if c.Name == "" || c.Image == "" {
			return false
		}
	}
	return true
}

func (kubelet *Kubelet) PullFromApiserver() {
	fmt.Println("Kubelet PullFromApiserver")

	podService := rtm.NewRemoteRuntimeService(time.Minute)
	defer podService.Close()

	podStatus, err := podService.GetAllPodsStatusOnNode() // 只返回了metadata和status，没有spec，spec要从api-server获取
	if err != nil {
		panic(err)
	}
	kubelet.Pods = kubelet.Pods[:0] // 清空以保证实时性（重启kubelet不影响pod运行）
	for _, p := range podStatus {
		kubelet.Pods = append(kubelet.Pods, *p) // 此时kubelet.Pods不包含spec
	}

	req, _ := json.Marshal("pod")
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	var podsFromApiserver []protocol.Pod
	err = json.Unmarshal(resp, &podsFromApiserver)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var pods []protocol.Pod
	for _, p := range podsFromApiserver {
		if !CheckPodSpec(p.Config) {
			fmt.Println("CheckPodSpec failed!")
			continue
		}
		pods = append(pods, p)
	}

	// 补充kubelet缺少的pod
	for _, p := range pods {
		if p.Config.NodeName != kubelet.Config.Name { // 保证是本节点的pod
			continue
		}
		existed := false
		for i, kp := range kubelet.Pods {
			if kp.Config.Metadata.Name == p.Config.Metadata.Name && kp.Config.Metadata.Namespace == p.Config.Metadata.Namespace {
				existed = true
				kubelet.Pods[i].Config = p.Config // 补充spec
				break
			}
		}
		if !existed {
			fmt.Println("Create pod from apiserver: ", p.Config.Metadata.Name, p.Config.Metadata.Namespace)
			podService.CreatePod(&p)
			kubelet.Pods = append(kubelet.Pods, p)
		}

	}

	// 删除kubelet多余的pod
	tmpPods := kubelet.Pods
	kubelet.Pods = kubelet.Pods[:0]
	for _, kp := range tmpPods {
		existed := false
		for _, p := range pods {
			if p.Config.NodeName != kubelet.Config.Name { // 保证是本节点的pod
				continue
			}
			if kp.Config.Metadata.Name == p.Config.Metadata.Name && kp.Config.Metadata.Namespace == p.Config.Metadata.Namespace {
				existed = true
				break
			}
		}
		if !existed {
			fmt.Println("Delete pod not in apiserver: ", kp.Config.Metadata.Name, kp.Config.Metadata.Namespace)
			podService.RemovePodSandBox(&kp)
		} else {
			kubelet.Pods = append(kubelet.Pods, kp)
		}
	}
}
