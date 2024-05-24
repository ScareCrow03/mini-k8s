package controller

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"time"
)

type ReplicasetController struct {
}

func (rsc *ReplicasetController) Start() {
	ticker := time.NewTicker(10 * time.Second)
	// defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				rsc.CheckAllReplicaset()
			}
		}
	}()

	go message.Consume(message.DeleteReplicasetQueueName, rsc.DeleteReplicaset)
	go message.Consume(message.ReplicasetCheckNowQueueName, rsc.ReplicasetCheckNow) // 解决延迟问题
}

// pod数量不足，创建pod，需要给pod name加随机5位后缀
func (rsc *ReplicasetController) CreatePod(rs protocol.ReplicasetType, num int, namespace string) {
	// fmt.Println("CreatePod: ", num)
	for range num {
		podConfig := rs.Config.Spec.Template
		podConfig.Metadata.Name = podConfig.Metadata.Name + "-" + uid.NewUid()[:5]
		podConfig.Metadata.Namespace = namespace

		req, err := json.Marshal(podConfig)
		if err != nil {
			fmt.Println("marshal request body failed")
			return
		}
		httputils.Post(constant.HttpPreffix+"/createPodFromFile", req)
	}
}

func (rsc *ReplicasetController) DeletePod(pods []protocol.Pod, num int, namespace string) {
	fmt.Println("DeletePod: ", num)
	for i := range num {
		pods[i].Config.Metadata.Namespace = namespace
		req, err := json.Marshal(pods[i].Config)
		if err != nil {
			fmt.Println("marshal request body failed")
			return
		}
		httputils.Post(constant.HttpPreffix+"/deletePodFromFile", req)
	}
}

func (rsc *ReplicasetController) CheckAllReplicaset() {
	fmt.Println("CheckAllReplicaset")
	req, _ := json.Marshal("replicaset")
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	var rss []protocol.ReplicasetType
	err := json.Unmarshal(resp, &rss)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// fmt.Println(len(rss))

	req, _ = json.Marshal("pod")
	resp = httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	var pods []protocol.Pod
	err = json.Unmarshal(resp, &pods)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// fmt.Println(len(pods))

	for _, rs := range rss {
		rs.Config.Spec.Selector.MatchLabels["ReplicasetMetadata"] = rs.Config.Metadata.Namespace + "/" + rs.Config.Metadata.Name
		ps := protocol.SelectPodsByLabelsNoPointer(rs.Config.Spec.Selector.MatchLabels, pods)

		fmt.Println("replicaset: ", rs.Config.Metadata.Namespace+"/"+rs.Config.Metadata.Name)
		// fmt.Println("pods: ", len(ps))
		if rs.Config.Spec.Replicas == len(ps) {
			continue
		}
		if rs.Config.Spec.Replicas > len(ps) {
			rsc.CreatePod(rs, rs.Config.Spec.Replicas-len(ps), rs.Config.Metadata.Namespace)
		} else {
			rsc.DeletePod(ps, len(ps)-rs.Config.Spec.Replicas, rs.Config.Metadata.Namespace)
		}
	}
}

// 删除所有pod
func (rsc *ReplicasetController) DeleteReplicaset(msg map[string]interface{}) error {
	fmt.Println("consume: " + message.DeleteReplicasetQueueName)
	var rs protocol.ReplicasetType
	rsjson, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	json.Unmarshal(rsjson, &rs.Config)

	req, _ := json.Marshal("pod")
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	var pods []protocol.Pod
	err = json.Unmarshal(resp, &pods)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	rs.Config.Spec.Selector.MatchLabels["ReplicasetMetadata"] = rs.Config.Metadata.Namespace + "/" + rs.Config.Metadata.Name
	ps := protocol.SelectPodsByLabelsNoPointer(rs.Config.Spec.Selector.MatchLabels, pods)
	rsc.DeletePod(ps, len(ps), rs.Config.Metadata.Namespace)

	return nil
}

func (rsc *ReplicasetController) ReplicasetCheckNow(msg map[string]interface{}) error {
	fmt.Println("consume: " + message.ReplicasetCheckNowQueueName)
	rsc.CheckAllReplicaset()
	httputils.Post(constant.HttpPreffix+"/serviceCheckNow", nil)
	return nil
}
