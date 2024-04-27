package kubelet

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	yamlParse "mini-k8s/pkg/utils/yaml"
)

func (kubelet *Kubelet) init(path string) error {
	yamlParse.YAMLParse(&kubelet.Config, path)

	req, err := json.Marshal(kubelet.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(kubelet.Config.ApiServerAddress+"/kubelet/register", req)

	return err
}

func MsgParse(msg map[string]interface{}, ptr interface{}) error {
	jsonInfo, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("json marshal error")
		return err
	}
	json.Unmarshal(jsonInfo, ptr)
	return nil
}

func (kubelet *Kubelet) Start() {
	go message.Consume(message.CreatePodQueueName, func(msg map[string]interface{}) error {
		var pod protocol.Pod
		podjson, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("json marshal error")
			return err
		}
		json.Unmarshal(podjson, &pod.Config)
		err = CreatePod(&pod)
		if err != nil {
			return err
		}
		kubelet.Pods = append(kubelet.Pods, pod)
		return nil
	})
}
