package kubelet

import (
	"encoding/json"
	"mini-k8s/pkg/httputils"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"time"
)

// 获取kubelet状态并发送给api-server，包括pod状态
func (kubelet *Kubelet) SendHeartbeat() {
	kubelet.Runtime = time.Since(kubelet.StartTime)

	podService := rtm.NewRemoteRuntimeService(time.Minute)
	defer podService.Close()
	podStatus, err := podService.GetAllPodsStatusOnNode() // 只返回了metadata和status，没有spec
	if err != nil {
		panic(err)
	}
	for i, p := range kubelet.Pods {
		for _, ps := range podStatus {
			if p.Config.Metadata.Name == ps.Config.Metadata.Name && p.Config.Metadata.Namespace == ps.Config.Metadata.Namespace {
				p.Config.NodeName = kubelet.Config.Name
				p.Status = ps.Status
				p.Status.NodeName = kubelet.Config.Name
				kubelet.Pods[i] = p
				break
			}
		}
	}

	// fmt.Println("SendHeartBeat, pods:")
	// for i, p := range kubelet.Pods {
	// 	fmt.Println(i, p.Config.Metadata.Name, p.Config.Metadata.Namespace, p.Status.IP, p.Status.NodeName, p.Status.Phase, p.Status.Runtime, p.Status.UpdateTime)
	// }

	req, err := json.Marshal(kubelet)
	if err != nil {
		panic(err)
	}
	httputils.Post(kubelet.Config.ApiServerAddress+"/kubelet/heartbeat", req)
}
