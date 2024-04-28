package kubelet

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/httputils"
	yamlParse "mini-k8s/pkg/utils/yaml"
)

func (kubelet *Kubelet) Init(path string) error {
	yamlParse.YAMLParse(&kubelet.Config, path)

	req, err := json.Marshal(kubelet.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	fmt.Println(kubelet.Config.ApiServerAddress + "/kubelet/register")
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
