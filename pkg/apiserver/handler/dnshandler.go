package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	kubelet2 "mini-k8s/pkg/kubelet"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleDnsCreate(c *gin.Context) {
	var requestBody protocol.Dns
	var res []etcd.GetReply
	c.BindJSON(&requestBody)
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "create dns from file:" + requestBody.Spec.Host,
	// })
	if requestBody.Metadata.Namespace == "" {
		requestBody.Metadata.Namespace = "default"
	}

	fmt.Println("update dns to etcd")
	//检查master中是否已经创建了该dns
	pathstr := "/registry/dns/" + requestBody.Metadata.Namespace + "/" + requestBody.Metadata.Name + "/"
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	res, err = st.GetWithPrefix(pathstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if len(res) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dns already exists",
		})
		return
	}

	fmt.Println("get service ip from etcd")
	//查询dns中的每一个service，通过etcd获取到其clusterip，记录到dns结构中
	for i, p := range requestBody.Spec.Paths {
		servicepath := constant.EtcdServicePrefix + requestBody.Metadata.Namespace + "/" + p.ServiceName
		res, err = st.GetWithPrefix(servicepath)
		fmt.Println(servicepath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if len(res) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "service not found",
			})
			return
		}
		var service protocol.ServiceType
		json.Unmarshal([]byte(res[0].Value), &service)
		fmt.Println(service)
		requestBody.Spec.Paths[i].ServiceIp = service.Config.Spec.ClusterIP
	}

	fmt.Println(requestBody)
	//完善dns结构的元数据
	requestBody.Metadata.UUID = uid.NewUid()
	dnsjson, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//将dns信息写入etcd
	err = st.Put(pathstr, dnsjson)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	msg, _ := json.Marshal(requestBody)

	//将dns信息发送到dns队列,转发给dnscontroller，让其修改nginx配置
	message.Publish(message.CreateDnsQueueName, msg)
	c.JSON(http.StatusOK, gin.H{
		"message": "create dns success",
	})
}

func HandleDnsDelete(c *gin.Context) {
	var requestBody protocol.Dns
	c.BindJSON(&requestBody)
	if requestBody.Metadata.Namespace == "" {
		requestBody.Metadata.Namespace = "default"
	}
	pathstr := "/registry/dns/" + requestBody.Metadata.Namespace + "/" + requestBody.Metadata.Name + "/"
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	defer st.Close()
	res, err := st.GetWithPrefix(pathstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	if len(res) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dns not found",
		})
	}
	err = st.DelWithPrefix(pathstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	//将dns信息发送到dns队列,转发给dnscontroller，让其修改nginx配置
	msg, _ := json.Marshal(requestBody)
	message.Publish(message.DeleteDnsQueueName, msg)
	c.JSON(http.StatusOK, gin.H{
		"message": "delete dns success",
	})

}

func HandleUpdateHost(c *gin.Context) {
	var requestBody protocol.DnsMsg
	c.BindJSON(&requestBody)
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	defer st.Close()
	//将dnscontroller中返回的新增数据的结构发送给worker节点的kubeproxy
	fmt.Println(requestBody)
	dns := requestBody.Dns
	if dns.Metadata.Namespace == "" {
		dns.Metadata.Namespace = "default"
	}
	hostConfig := requestBody.HostConfig
	var sendmsg protocol.DnsMsg
	sendmsg.Dns = dns
	sendmsg.HostConfig = hostConfig
	sendmsgjson, err := json.Marshal(sendmsg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	//通过etcd找到所有的worker节点
	res, err := st.GetWithPrefix(constant.EtcdKubeletPrefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	for _, r := range res {
		var kubelet kubelet2.Kubelet
		json.Unmarshal([]byte(r.Value), &kubelet)
		fmt.Println(kubelet.Config.Name)
		fmt.Printf("Update host for svc %s\n", kubelet.Config.Name)
		message.Publish(message.UpdateHostQueueName+"-"+kubelet.Config.Name, sendmsgjson)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "update host config success",
	})
}
