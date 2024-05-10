package kubelet

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/httputils"
	yamlParse "mini-k8s/pkg/utils/yaml"
	"time"
)

func (kubelet *Kubelet) Init(path string) error {
	yamlParse.YAMLParse(&kubelet.Config, path)

	kubelet.StartTime = time.Now()
	kubelet.Runtime = time.Since(kubelet.StartTime)

	req, err := json.Marshal(kubelet.Config)
	if err != nil {
		panic(err)
	}
	fmt.Println(kubelet.Config.ApiServerAddress + "/kubelet/register")
	httputils.Post(kubelet.Config.ApiServerAddress+"/kubelet/register", req)

	return err
}
