package kubelet

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/httputils"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"time"
)

// 获取kubelet状态并发送给api-server，包括pod状态
func (kubelet *Kubelet) SendHeartbeat() {
	fmt.Println("Kubelet SendHeartbeat")
	fmt.Println("podsNum: ", len(kubelet.Pods))
	kubelet.Runtime = time.Since(kubelet.StartTime)
	// 更新kubelet的最后一次更新时间！这个时间是用来判断kubelet是否存活的
	kubelet.LastUpdateTime = time.Now()

	podService := rtm.NewRemoteRuntimeService(time.Minute)
	defer podService.Close()
	podStatus, err := podService.GetAllPodsStatusOnNode() // 只返回了metadata和status，没有spec
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for i, p := range kubelet.Pods {
		for _, ps := range podStatus {
			if p.Config.Metadata.Name == ps.Config.Metadata.Name && p.Config.Metadata.Namespace == ps.Config.Metadata.Namespace {
				// p.Config.NodeName = kubelet.Config.Name // 此项由scheduler写
				p.Status = ps.Status
				kubelet.Pods[i] = p
				// sss, err := json.Marshal(kubelet.Pods[i].Status.CtrsMetrics)
				// fmt.Println(string(sss))
				// if err != nil {
				// 	fmt.Println(err.Error())
				// }
				// sss, err = json.Marshal(kubelet.Pods[i].Status.PodMetrics)
				// fmt.Println(string(sss))
				// if err != nil {
				// 	fmt.Println(err.Error())
				// }
				break
			}
		}
	}
	// sss, _ := json.Marshal(kubelet.Pods)
	// fmt.Println(string(sss))

	for _, p := range kubelet.Pods {
		// fmt.Println("pod: ", p.Config.Metadata.Name, p.Config.Metadata.Namespace, p.Status.Phase, len(p.Status.ContainerStatus))
		_, otherCtrs, _ := podService.ListPodContainersById(p.Config.Metadata.UID)
		for _, c := range otherCtrs {
			s := p.Status.ContainerStatus[c.ID].Status
			// fmt.Println("container: ", s)
			if s == "dead" || s == "exited" {
				podService.StopContainer(c.ID)
				podService.StartContainer(c.ID)
			}
		}
	}

	// fmt.Println("SendHeartBeat, pods:")
	// for i, p := range kubelet.Pods {
	// 	fmt.Println(i, p.Config.Metadata.Name, p.Config.Metadata.Namespace, p.Status.IP, p.Config.NodeName, p.Status.Phase, p.Status.Runtime, p.Status.UpdateTime)
	// }

	req, err := json.Marshal(kubelet)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	httputils.Post(kubelet.Config.ApiServerAddress+"/kubelet/heartbeat", req)
}
