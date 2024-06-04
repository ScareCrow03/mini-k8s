package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	kubelet2 "mini-k8s/pkg/kubelet"
	"mini-k8s/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func KubeletRegister(c *gin.Context) {
	// TODO: register kubelet to apiserver, write into etcd
	var kubelet kubelet2.Kubelet
	// 最新更改为注册时把整个kubelet对象都传过来，即使现在没有Pods，但也有一些需要的动态字段
	c.BindJSON(&kubelet)
	fmt.Println(kubelet.Config.Name)
	kubelet.Pods = nil
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	jsonstr, err := json.Marshal(kubelet)
	if err != nil {
		panic(err)
	}
	st.Put(constant.EtcdKubeletPrefix+kubelet.Config.Name, jsonstr)

	c.JSON(http.StatusOK, "kubelet register: "+kubelet.Config.Name)
}

func KubeletHeartbeat(c *gin.Context) {
	var kubelet kubelet2.Kubelet
	c.BindJSON(&kubelet)
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	// 以下把kubelet自身的信息写入etcd的kubelet路径，即拷贝一份相应的元数据即可，此时不需要Pods了
	var kubeletInEtcd kubelet2.Kubelet
	reply2, err := st.Get(constant.EtcdKubeletPrefix + kubelet.Config.Name)
	if err != nil {
		logger.KError("etcd get error: %v", err)
		c.JSON(http.StatusOK, nil)
		return
	}
	if len(reply2.Value) == 0 { // 任何原因没有找到它
		logger.KError("node %s not found in etcd, but get its heartbeat!", kubelet.Config.Name)
		c.JSON(http.StatusOK, nil)
		return
	}

	json.Unmarshal(reply2.Value, &kubeletInEtcd)
	// 只更新必要的字段
	kubeletInEtcd.LastUpdateTime = kubelet.LastUpdateTime
	kubeletInEtcd.Runtime = kubelet.Runtime
	// kubeletInEtcd.Status = kubelet.Status
	// kubeletInEtcd.Pods = kubelet.Pods

	jsonstr2, _ := json.Marshal(kubeletInEtcd)
	st.Put(constant.EtcdKubeletPrefix+kubelet.Config.Name, jsonstr2)

	// 检验etcd中的pod信息与kubelet heartbeat是否相符，心跳中只能包含etcd中存在的pod的信息

	// 心跳中只能包含etcd中存在的pod的信息
	for _, p := range kubelet.Pods {

		jsonstr, err := json.Marshal(p)
		if err != nil {
			panic(err)
		}
		reply, err := st.Get(constant.EtcdPodPrefix + p.Config.Metadata.Namespace + "/" + p.Config.Metadata.Name)
		if err != nil {
			panic(err)
		}
		if len(reply.Value) > 0 {
			st.Put(constant.EtcdPodPrefix+p.Config.Metadata.Namespace+"/"+p.Config.Metadata.Name, jsonstr)
		}
	}

	c.JSON(http.StatusOK, nil)

}
