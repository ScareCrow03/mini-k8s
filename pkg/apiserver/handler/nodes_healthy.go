package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	kubelet2 "mini-k8s/pkg/kubelet"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	"time"
)

const (
	NodesExitTimeLimit = 25 * time.Second
)

func CheckNodesHealthy() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	st, _ := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	defer st.Close()
	// 每15s检查一次kubelet是否还存活；判断是否存活的标准，根据Kubelet上一次发心跳的时间，到现在的距离测量是否超过30s；为什么循环检查时间与判断时间不一致？因为如果一致，可能0s有一个kubelet发心跳退出、然后29s有一个检查没有超过30s，则需要2倍的时间才能发现
	for {
		select {
		case <-ticker.C:
			fmt.Printf("CheckNodesHealthy\n")
			// 遍历当前所有kubelet，检查是否还存活
			allNodes, err := st.GetWithPrefix(constant.EtcdKubeletPrefix)
			if err != nil {
				logger.KError("etcd get error: %v", err)
				continue
			}

			// 遍历所有Pods，将部署在该kubelet上的Pods的Phase标注为Unknown
			allPods, err := st.GetWithPrefix(constant.EtcdPodPrefix)
			if err != nil {
				logger.KError("etcd get error: %v", err)
				continue
			}

			for _, r := range allNodes {
				var kubelet kubelet2.Kubelet
				err = json.Unmarshal(r.Value, &kubelet)
				if err != nil {
					logger.KError("json unmarshal error: %v", err)
					continue
				}

				// 如果该kubelet已经超过20秒没有更新，我们认为它已经退出
				if time.Since(kubelet.LastUpdateTime) > NodesExitTimeLimit {
					logger.KInfo("nodeName %s, nodeIP %s is not healthy, delete it and set related Pods Phase Unknown", kubelet.Config.Name, kubelet.Config.NodeIP)
					st.Del(constant.EtcdKubeletPrefix + kubelet.Config.Name)

					// 标注该kubelet上的Pods状态为Unknown，具体为查看各pods的静态字段NodeName，然后与kubelet.Config.Name比较
					for _, r := range allPods {
						var pod protocol.Pod
						err = json.Unmarshal(r.Value, &pod)
						if err != nil {
							logger.KError("json unmarshal error: %v", err)
						}

						// 建议比较静态字段的NodeName，而不是status！
						if pod.Config.NodeName == kubelet.Config.Name {
							// 只需要改动状态字段为Unknown，再写回去即可
							pod.Status.Phase = constant.PodPhaseUnknown
							jsonstr, err := json.Marshal(pod)
							if err != nil {
								logger.KError("json marshal error: %v", err)
							}
							st.Put(constant.EtcdPodPrefix+pod.Config.Metadata.Namespace+"/"+pod.Config.Metadata.Name, jsonstr)
						}
					}
				}
			}
		}
	}
}
