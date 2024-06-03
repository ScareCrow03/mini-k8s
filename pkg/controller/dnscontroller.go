package controller

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/nginx"
	yamlParse "mini-k8s/pkg/utils/yaml"
	"os/exec"
	"time"
)

type DnsController struct {
	NginxServiceIp   string
	HostList         []string
	NginxServiceName string
	ContainerID      string
}

var Dc = DnsController{}

func Init() {
	//创建一个nginx pod
	var nginxPod protocol.Pod
	yamlParse.YAMLParse(&nginxPod.Config, constant.NginxPodPath)
	req, err := json.Marshal(nginxPod.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return
	}
	httputils.Post("http://localhost:8080/createPodFromFile", req)

	//等待5s
	time.Sleep(5 * time.Second)
	GetNginxContainerId()

	//创建一个nginx service
	var nginxService protocol.ServiceType
	yamlParse.YAMLParse(&nginxService.Config, constant.NginxServicePath)
	req, err = json.Marshal(nginxService)
	if err != nil {
		fmt.Println("marshal request body failed")
		return
	}
	rep := httputils.Post("http://localhost:8080/createServiceFromFile", req)

	var service protocol.ServiceType
	fmt.Println(rep)
	err = json.Unmarshal(rep, &service)
	fmt.Println("ww")
	if err != nil {
		fmt.Println("err")
		fmt.Println(err.Error())
		return
	}
	//通过post的返回值得到该service的ip地址，记录到自己的数据结构中
	Dc.NginxServiceIp = service.Config.Spec.ClusterIP
	fmt.Println("asdfas")
	//监听由apiserver发来的dns创建请求或修改请求
	go message.Consume(message.CreateDnsQueueName, handleCreateDns)
	go message.Consume(message.DeleteDnsQueueName, handleDeleteDns)
	fmt.Println("asdfass")
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

	nginx.WriteNginxConf(dns)
	//重新生成nginx pod, 可能无法重启
	// var pod1 protocol.Pod
	// yamlParse.YAMLParse(&pod1.Config, constant.NginxPodPath)
	// req, err = json.Marshal(pod1.Config)
	// if err != nil {
	// 	fmt.Println("marshal request body failed")
	// 	return err
	// }
	// httputils.Post(constant.HttpPreffix+"/deletePodFromFile", req)
	// //等待2s
	// time.Sleep(2 * time.Second)
	// httputils.Post(constant.HttpPreffix+"/createPodFromFile", req)

	// 直接找到该docker容器，让其执行nginx -s reload
	cmd := "docker exec " + Dc.ContainerID + " sh -c 'nginx -s reload'"
	fmt.Println(cmd)
	exec.Command("bash", "-c", cmd).Run()

	var dnsmsg protocol.DnsMsg
	dnsmsg.Dns = dns
	dnsmsg.HostConfig = Dc.HostList
	sendmsgjson, err := json.Marshal(dnsmsg)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post("http://localhost:8080/updateHost", sendmsgjson)
	return nil
}

func handleDeleteDns(msg map[string]interface{}) error {
	fmt.Printf("handleDeleteDns: %s\n", msg)
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
	//根据发来的dns消息，修改host映射, 删除dns对应的条目
	for i, host := range Dc.HostList {
		if host == Dc.NginxServiceIp+" "+dns.Spec.Host {
			Dc.HostList = append(Dc.HostList[:i], Dc.HostList[i+1:]...)
			break
		}
	}
	nginx.DeleteNginxConf(dns)
	//重新生成nginx pod
	var pod1 protocol.Pod
	yamlParse.YAMLParse(&pod1.Config, constant.NginxPodPath)
	req, err = json.Marshal(pod1.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/deletePodFromFile", req)
	//等待2s
	time.Sleep(2 * time.Second)
	httputils.Post(constant.HttpPreffix+"/createPodFromFile", req)
	var dnsmsg protocol.DnsMsg
	dnsmsg.Dns = dns
	dnsmsg.HostConfig = Dc.HostList
	sendmsgjson, err := json.Marshal(dnsmsg)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post("http://localhost:8080/updateHost", sendmsgjson)
	return nil
}

func GetNginxContainerId() {
	cmd := "docker ps | grep nginx | awk '{print $1}'"
	command := exec.Command("bash", "-c", cmd)
	output, err := command.Output()
	if err != nil {
		fmt.Println("get nginx container id failed")
		return
	}
	fmt.Println("conatinerid is:", string(output))
	if len(output) == 0 {
		fmt.Println("get nginx container id failed")
		return
	}
	//删除ouput后面的换行符
	output = output[:len(output)-1]
	Dc.ContainerID = string(output)
}
