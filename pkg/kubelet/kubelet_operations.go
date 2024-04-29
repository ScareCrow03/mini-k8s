package kubelet

import (
	"fmt"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"time"
)

func CreatePod(pod *protocol.Pod) error {
	podService := rtm.NewRemoteRuntimeService(time.Minute)
	defer podService.Close()
	// fmt.Println(pod.Config.Metadata.Name, pod.Config.Metadata.Namespace)
	err := podService.CreatePod(pod)
	if err != nil {
		panic(err)
	}
	err = podService.StartPod(pod)
	if err != nil {
		panic(err)
	}
	return nil
}

func StopPod(pod *protocol.Pod) error {
	podService := rtm.NewRemoteRuntimeService(time.Minute)
	defer podService.Close()
	err := podService.StopPodSandBox(pod)
	if err != nil {
		panic(err)
	}
	return nil
}

func DeletePod(pod *protocol.Pod) error {
	podService := rtm.NewRemoteRuntimeService(time.Minute)
	defer podService.Close()
	fmt.Println(pod.Config.Metadata.Name, pod.Config.Metadata.Namespace)
	err := podService.StopPodSandBox(pod)
	if err != nil {
		panic(err)
	}
	err = podService.RemovePodSandBox(pod)
	if err != nil {
		panic(err)
	}
	return nil
}
