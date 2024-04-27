package kubelet

import (
	"fmt"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"time"
)

func CreatePod(pod *protocol.Pod) error {
	podService := rtm.NewRemoteRuntimeService(5 * time.Minute)
	err := podService.CreatePod(pod)
	if err != nil {
		fmt.Printf("Failed to create pod: %v", err)
		return err
	}
	return nil
}
