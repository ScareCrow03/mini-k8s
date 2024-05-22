package controller

import (
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	"time"

	"encoding/json"

	"gopkg.in/yaml.v3"
)

type HPAController struct {
	// 认为HPA对象创建后只读，做更新之后它的UID一定发生变化
	HpasMap map[string]*protocol.HPAType
	// HPA_ID -> Ticker的映射，用于定期检查某个HPA对象
	Tickers map[string]*time.Ticker
	// 新增一个map来存储每个HPA对象的退出通道
	QuitChs map[string]chan struct{}
}

func (hpaC *HPAController) Start() {
	hpaC.HpasMap = make(map[string]*protocol.HPAType)
	hpaC.Tickers = make(map[string]*time.Ticker)
	hpaC.QuitChs = make(map[string]chan struct{})
	// 建议比replicaSet处理时间长一点
	ticker := time.NewTicker(15 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				hpaC.CheckAllHPA()
			}
		}
	}()
}

// 增量式地同步HPA对象
func (hpaC *HPAController) CheckAllHPA() {
	fmt.Printf("CheckAllHPA\n")
	req, _ := json.Marshal("hpa")
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	// 获取所有hpa对象
	var hpas []protocol.HPAType
	err := json.Unmarshal(resp, &hpas)
	if err != nil {
		logger.KError("unmarshal hpas error %s", err)
		return
	}

	updatedHpas := make(map[string]*protocol.HPAType)
	for _, hpa := range hpas {
		// 用UID唯一标识一个HPA对象
		hpaKey := hpa.Config.Metadata.UID
		updatedHpas[hpaKey] = &hpa

		// 如果是新的HPA对象，创建新的定时器和退出通道
		if _, ok := hpaC.HpasMap[hpaKey]; !ok {
			data, _ := yaml.Marshal(hpa)
			fmt.Printf("Detect new HPA created, create ticker for it: \n%s\n,", string(data))
			hpaC.Tickers[hpaKey] = time.NewTicker(time.Duration(hpa.Config.Spec.ScaleInterval) * time.Second)
			// 退出通道用于释放资源，需要良好维护这个句柄
			hpaC.QuitChs[hpaKey] = make(chan struct{})
			go func(hpa protocol.HPAType) {
				// 创建后立即先检查一次，否则等待时间太长了！
				hpaC.CheckOneHPA(hpa)
				for {
					select {
					case <-hpaC.Tickers[hpaKey].C:
						hpaC.CheckOneHPA(hpa) // 检查单个HPA对象
					case <-hpaC.QuitChs[hpaKey]:
						return // 收到退出信号，结束协程
					}
				}
			}(hpa)
			hpaC.HpasMap[hpaKey] = &hpa
		}
		// 已经有的HPA对象完全不用管，因为要求UID唯一确定只读对象

	}

	// 对于HpasMap中有但是在新的HPA列表中没有的，停止并删除对应的定时器和退出通道
	for hpaKey, ticker := range hpaC.Tickers {
		if _, ok := updatedHpas[hpaKey]; !ok {
			data, _ := yaml.Marshal(hpaC.HpasMap[hpaKey])
			fmt.Printf("Detect current HPA removed, stop ticker for it: %s", string(data))
			ticker.Stop()
			close(hpaC.QuitChs[hpaKey]) // 关闭退出通道，通知协程退出
			delete(hpaC.Tickers, hpaKey)
			delete(hpaC.QuitChs, hpaKey)
			delete(hpaC.HpasMap, hpaKey)
		}
	}
}

// 这个函数用于在一个ticker协程里，定期检查一个HPA对象；需要各自维护状态
func (hpaC *HPAController) CheckOneHPA(hpa protocol.HPAType) {
	fmt.Printf("CheckOneHPA %s/%s, minReplica: %v, maxReplica: %v, scaleInterval: %v, targetReplicaSetName: %s\n", hpa.Config.Metadata.Namespace, hpa.Config.Metadata.Name, hpa.Config.Spec.MinReplicas, hpa.Config.Spec.MaxReplicas, hpa.Config.Spec.ScaleInterval, hpa.Config.Spec.ScaleTargetRef.Name)
	// 获取所有pod对象
	req, _ := json.Marshal("pod")
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	var pods []protocol.Pod
	err := json.Unmarshal(resp, &pods)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 根据hpa的配置，获得replicaSet的元数据索引
	rsMeta := protocol.Metadata{
		Namespace: hpa.Config.Metadata.Namespace,
		Name:      hpa.Config.Spec.ScaleTargetRef.Name,
	}
	req2, _ := json.Marshal(rsMeta)
	// 尝试找到对应的replicaset
	resp2 := httputils.Post(constant.HttpPreffix+"/getOneReplicaset", req2)
	// 获取单个对象，进行一些详尽的检查，保证只有在成功获取后才进行下一步操作；如果没有对应的replicaSet静态信息，那么跳过
	if resp2 == nil {
		return
	}

	var rs protocol.ReplicasetType
	err = json.Unmarshal(resp2, &rs)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if rs.Config.Metadata.Name == "" {
		return
	}

	// 现在已经获取到了replicaset，接下来需要获取到replicaset下的所有pod
	rs.Config.Spec.Selector.MatchLabels["ReplicasetMetadata"] = rs.Config.Metadata.Namespace + "/" + rs.Config.Metadata.Name
	// 筛选出replicaset下的所有pod
	managedPods := protocol.SelectPodsByLabelsNoPointer(rs.Config.Spec.Selector.MatchLabels, pods)

	// 计算replicaset级别的所有pod的资源利用情况，只是做一个简单的算术平均
	podsMetricsEntries := make(map[string]protocol.PodMetricsEntry, 0)
	for _, pod := range managedPods {
		podsMetricsEntries[pod.Config.Metadata.UID] = pod.Status.PodMetrics
	}
	rsMetricsResult := protocol.CalculateReplicaMetrics(&hpa, podsMetricsEntries)

	// 当前replicaset的副本数，应该取spec静态期望值，而非真实值，可以留给replicaSetController一些调整时间
	currentReplicaNum := rs.Config.Spec.Replicas
	// 期望副本数
	desiredReplicaNum := protocol.CalculateDesiredReplicas(&hpa, currentReplicaNum, rsMetricsResult)

	// 比较
	if currentReplicaNum == desiredReplicaNum {
		// 如果已经相等，则不管
		return
	} else {
		if currentReplicaNum < desiredReplicaNum {
			// 如果当前副本数小于期望副本数，则副本数自增1
			rs.Config.Spec.Replicas = currentReplicaNum + 1
		} else {
			rs.Config.Spec.Replicas = currentReplicaNum - 1
		}

		// 发给api-server，为这个replicaSet应用新的副本数；注意这里直接使用create方法，传递相同的静态配置（不需要管UID的问题，重建一份也没事），只是副本数不同
		req3, err := json.Marshal(rs.Config)
		if err != nil {
			logger.KError("marshal request body failed，err: %s", err)
			return
		}
		httputils.Post(constant.HttpPreffix+"/createReplicasetFromFile", req3)
	}
}
