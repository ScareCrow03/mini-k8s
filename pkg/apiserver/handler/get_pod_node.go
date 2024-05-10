package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/protocol"
)

func GetPodNode(podConfig protocol.PodConfig) string {
	etcdStore, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	defer etcdStore.Close()
	reply, err2 := etcdStore.Get(constant.EtcdPodPrefix + podConfig.Metadata.Namespace + "/" + podConfig.Metadata.Name)
	if err2 != nil {
		fmt.Println(err2.Error())
		return ""
	}
	var pod protocol.Pod
	err = json.Unmarshal(reply.Value, &pod)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return pod.Config.NodeName
}
