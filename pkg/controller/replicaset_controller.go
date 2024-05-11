package controller

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"time"
)

type ReplicasetController struct {
}

func (rsc *ReplicasetController) Start() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				rsc.CheckAllReplicaset()
			}
		}
	}()
}

// pod数量不足，创建pod，需要给pod name加随机5位后缀
func (rsc *ReplicasetController) CreatePod(rs protocol.ReplicasetType, num int) {
	for range num {
		podConfig := rs.Config.Spec.Template
		podConfig.Metadata.Name = podConfig.Metadata.Name + "-" + uid.NewUid()[:5]

		req, err := json.Marshal(podConfig)
		if err != nil {
			fmt.Println("marshal request body failed")
			return
		}
		httputils.Post(constant.HttpPreffix+"/createPodFromFile", req)
	}
}

func (rsc *ReplicasetController) DeletePod(pods []protocol.Pod, num int) {
	for i := range num {
		req, err := json.Marshal(pods[i].Config)
		if err != nil {
			fmt.Println("marshal request body failed")
			return
		}
		httputils.Post(constant.HttpPreffix+"/deletePodFromFile", req)
	}
}

func (rsc *ReplicasetController) CheckAllReplicaset() {
	req, _ := json.Marshal("replicaset")
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	var rss []protocol.ReplicasetType
	err := json.Unmarshal(resp, &rss)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	req, _ = json.Marshal("pod")
	resp = httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	var pods []protocol.Pod
	err = json.Unmarshal(resp, &rss)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, rs := range rss {
		rs.Config.Metadata.Labels["ReplicasetMetadata"] = rs.Config.Metadata.Namespace + "/" + rs.Config.Metadata.Name
		ps := protocol.SelectPodsByLabelsNoPointer(rs.Config.Metadata.Labels, pods)
		if rs.Config.Spec.Replicas == len(ps) {
			continue
		}
		if rs.Config.Spec.Replicas > len(ps) {
			rsc.CreatePod(rs, rs.Config.Spec.Replicas-len(ps))
		} else {
			rsc.DeletePod(ps, len(ps)-rs.Config.Spec.Replicas)
		}
	}
}
