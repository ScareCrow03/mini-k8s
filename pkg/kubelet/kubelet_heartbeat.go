package kubelet

import (
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"time"
)

func (kubelet Kubelet) SendPodHeartbeat() {
	podService := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer podService.Close()
	// for _, p := range kubelet.Pods {
	// 	podStatus, err := podService.PodSandBoxStatus(&p)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// }
}
