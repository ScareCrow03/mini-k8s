package kubelet

import (
	"fmt"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"time"
)

func (kubelet *Kubelet) CreatePod(pod *protocol.Pod) error {
	podService := rtm.NewRemoteRuntimeService(time.Minute)
	defer podService.Close()
	// fmt.Println(pod.Config.Metadata.Name, pod.Config.Metadata.Namespace)
	err := podService.CreatePod(pod)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = podService.StartPod(pod)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (kubelet *Kubelet) StopPod(pod *protocol.Pod) error {
	podService := rtm.NewRemoteRuntimeService(time.Minute)
	defer podService.Close()
	err := podService.StopPodSandBox(pod)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (kubelet *Kubelet) DeletePod(pod *protocol.Pod) error {
	podService := rtm.NewRemoteRuntimeService(time.Minute)
	defer podService.Close()
	fmt.Println(pod.Config.Metadata.Name, pod.Config.Metadata.Namespace)
	err := podService.StopPodSandBox(pod)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = podService.RemovePodSandBox(pod)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
