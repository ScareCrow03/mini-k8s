package kubelet

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/utils/net_util"
	yamlParse "mini-k8s/pkg/utils/yaml"
	"time"
)

func (kubelet *Kubelet) Init(path string) error {
	yamlParse.YAMLParse(&kubelet.Config, path)

	kubelet.StartTime = time.Now()
	kubelet.Runtime = time.Since(kubelet.StartTime)
	kubelet.LastUpdateTime = time.Now()

	// 注册时把NodeIP也传过去
	kubelet.Config.NodeIP, _ = net_util.GetNodeIP()

	kubelet.PullFromApiserver()

	// 请把整个kubelet的信息都传给api-server！
	req, err := json.Marshal(kubelet)
	if err != nil {
		panic(err)
	}
	fmt.Println(kubelet.Config.ApiServerAddress + "/kubelet/register")
	httputils.Post(kubelet.Config.ApiServerAddress+"/kubelet/register", req)

	return err
}
