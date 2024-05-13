package controller

import (
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	"time"

	"encoding/json"
)

type HPAController struct {
}

// hpaController自己不需要维护任何状态，只需要定期（例如每10秒）从API服务器获取所有的HPA对象，然后根据每个HPA的配置找到对应的ReplicaSet以及其Pods，并根据当前的资源利用情况和HPA的配置来决定是否需要进行扩缩容。如果desired和replicaSet的静态副本信息相同（而不是去看实际的Pods数，因为完全可能replicaSet静态信息已经到位，但是因为ReplicaSet自己间隔比较长、还没有来得及处理最新的replicaSet信息、也就没有创建新的Pod信息，这样只需要等一会即可），那么无需再申请重建这个replicaSet
// 这与kube-proxy不同，因为kube-proxy希望做ipvs增量式的更新，那么必然要具体到每一个节点上独特的状态
// 如果hpa对象存在，但是对应的replicaSet静态信息没有，简单处理为跳过这个hpa对象的相关计算
// 观察到，replicaSet的删除信息需要controller来手动消费，否则api-server只是删除了etcd中的静态数据，而没有管理底下的几个pods；但是新增的消息可以定期拉取到，这个是OK的；但是hpa的删除信息并不需要controller来消费，因为被删除后下一轮就不会再被拉取到了，也就不会有动态监测资源并做出调整的逻辑，这是OK的
func (hpaC *HPAController) Start() {
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

func (hpaC *HPAController) CheckAllHPA() {
	fmt.Println("CheckAllHPA")
	req, _ := json.Marshal("HPA")
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	// 获取所有hpa对象
	var hpas []protocol.HPAType
	err := json.Unmarshal(resp, &hpas)
	if err != nil {
		logger.KError("unmarshal hpas error %s", err)
		return
	}
	// 获取所有pod对象
	req, _ = json.Marshal("pod")
	resp = httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	var pods []protocol.Pod
	err = json.Unmarshal(resp, &pods)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 遍历所有hpa对象，根据hpa查看是否需要处理该replicaset
	for _, hpa := range hpas {
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
			continue
		}
		var rs protocol.ReplicasetType
		err = json.Unmarshal(resp2, &rs)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if rs.Config.Metadata.Name == "" {
			continue
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
			continue
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
}
