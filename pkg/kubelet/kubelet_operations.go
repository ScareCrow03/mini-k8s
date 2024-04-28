package kubelet

import (
	"fmt"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"time"
)

func CreatePod(pod *protocol.Pod) error {
	podService := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer podService.Close()
	// fmt.Println(pod.Config.Metadata.Name, pod.Config.Metadata.Namespace)
	err := podService.CreatePod(pod)
	if err != nil {
		fmt.Printf("Failed to create pod: %v", err)
		return err
	}
	err = podService.StartPod(pod)
	if err != nil {
		fmt.Printf("Failed to start pod: %v", err)
		return err
	}
	return nil
}

func StopPod(pod *protocol.Pod) error {
	podService := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer podService.Close()
	err := podService.StopPodSandBox(pod)
	if err != nil {
		fmt.Printf("Failed to stop pod: %v", err)
		return err
	}
	return nil
}

func DeletePod(pod *protocol.Pod) error {
	podService := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer podService.Close()
	fmt.Println(pod.Config.Metadata.Name, pod.Config.Metadata.Namespace)
	defer podService.Close()
	err := podService.StopPodSandBox(pod)
	if err != nil {
		fmt.Printf("Failed to stop pod: %v", err)
		return err
	}
	err = podService.RemovePodSandBox(pod)
	if err != nil {
		fmt.Printf("Failed to delete pod: %v", err)
		return err
	}
	return nil
}

// TODO get pod status
// func GetPod(pod *protocol.Pod) error {
// 	podService := rtm.NewRemoteRuntimeService(5 * time.Minute)
//  defer podService.Close()
// 	// 查看本pod状态
// 	podStatus, err = podService.GetPodStatusById(pod.Config.Metadata.UID)
// 	if err != nil {
// 		fmt.Printf("Failed to get pod: %v", err)
// 		return err
// 	}
// }
