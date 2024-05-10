package dns_server

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"os"

	"mini-k8s/pkg/utils/nginx"
	yamlParse "mini-k8s/pkg/utils/yaml"
)

func init() {
	//创建一个nginx pod
	var nginxPod protocol.Pod
	yamlParse.YAMLParse(&nginxPod.Config, constant.NginxPodPath)
	req, err := json.Marshal(nginxPod.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return
	}
	httputils.Post("http://localhost:8080/createPodFromFile", req)

	//监听由apiserver发来的dns创建请求或修改请求
	go message.Consume(message.UpdateHostQueueName, handleUpdateHostmsg)
}

func handleUpdateHostmsg(msg map[string]interface{}) error {
	fmt.Printf("handleUpdateHostmsg: %s\n", msg)
	hostConfig := msg["hostconfig"].([]string)
	dns := msg["dns"].(protocol.Dns)
	//根据发来的dns消息，修改host映射
	file, err := os.OpenFile(constant.HostsFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer file.Close()
	for _, host := range hostConfig {
		file.WriteString(host + "\n")
	}
	nginx.WriteNginxConf(dns)
	return nil
}
