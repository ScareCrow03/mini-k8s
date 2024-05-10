package controller

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	yamlParse "mini-k8s/pkg/utils/yaml"
)

type DnsController struct {
	NginxServiceIp   string
	HostList         []string
	NginxServiceName string
}

var Dc = DnsController{}

func init() {
	//创建一个nginx service
	var nginxService protocol.ServiceType
	yamlParse.YAMLParse(&nginxService, constant.NginxServicePath)
	req, err := json.Marshal(nginxService)
	if err != nil {
		fmt.Println("marshal request body failed")
		return
	}
	httputils.Post("http://localhost:8080/createServiceFromFile", req)

	//通过post的返回值得到该service的ip地址，记录到自己的数据结构中
	// Dc.NginxServiceIp = ""
	//监听由apiserver发来的dns创建请求或修改请求
	go message.Consume(message.CreateDnsQueueName, handleCreateDns)
	// go message.Consume(message.UpdateDnsQueueName, handleUpdateDns)
}

func handleCreateDns(msg map[string]interface{}) error {
	fmt.Printf("handleCreateDns: %s\n", msg)
	var dns protocol.Dns
	req, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	err = json.Unmarshal(req, &dns)
	if err != nil {
		fmt.Println("unmarshal request body failed")
		return err
	}
	//根据发来的dns消息，修改host映射
	Dc.HostList = append(Dc.HostList, Dc.NginxServiceIp+" "+dns.Spec.Host)

	sendmsg := map[string]interface{}{
		"dns":        dns,
		"hostconfig": Dc.HostList,
	}
	sendmsgjson, err := json.Marshal(sendmsg)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post("http://localhost:8080/updateHost", sendmsgjson)
	return nil
}
