package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllPods() []protocol.Pod {
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.GetWithPrefix(constant.EtcdPodPrefix)
	if err != nil {
		panic(err)
	}
	var pods []protocol.Pod
	for _, r := range reply {
		var p protocol.Pod
		err = json.Unmarshal(r.Value, &p)
		if err != nil {
			panic(err)
		}
		pods = append(pods, p)
	}
	return pods
}

func CreatePod(c *gin.Context) {
	var pod protocol.Pod
	c.BindJSON(&pod.Config)
	if pod.Config.Metadata.Namespace == "" {
		pod.Config.Metadata.Namespace = "default"
	}
	// 先检查是否有重复的资源，这里先检查一下pod
	// TODO: 检查namespace和name均相同的资源，不仅限于pod
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.Get(constant.EtcdPodPrefix + pod.Config.Metadata.Namespace + "/" + pod.Config.Metadata.Name)
	if err != nil {
		panic(err)
	}
	if len(reply.Value) > 0 { // 存在重复pod，阻止
		// msg, _ := json.Marshal(pod.Config)
		// nodeName := GetPodNode(pod.Config)
		// message.Publish(message.KubeletDeletePodQueue+"/"+nodeName, msg)

		// st.Del(constant.EtcdPodPrefix + pod.Config.Metadata.Namespace + "/" + pod.Config.Metadata.Name)
		fmt.Println("Create pod from file failed: same pod namespace & name")
		c.JSON(http.StatusOK, "Create pod from file failed: same pod namespace & name")
		return
	}

	msg, _ := json.Marshal(pod.Config)
	message.Publish(message.CreatePodQueueName, msg)

	c.JSON(http.StatusOK, "create pod from file: "+pod.Config.Metadata.Namespace+"/"+pod.Config.Metadata.Name)
}

func HandlePodAssignToNode(c *gin.Context) {
	var pod protocol.Pod
	c.BindJSON(&pod.Config)

	nodeName := pod.Config.NodeName
	pod.Config.Metadata.UID = "mini-k8s-pod-" + uid.NewUid()
	if pod.Config.Metadata.Namespace == "" {
		pod.Config.Metadata.Namespace = "default"
	}

	msg, _ := json.Marshal(pod.Config)
	message.Publish(message.KubeletCreatePodQueue+"/"+nodeName, msg)

	// 将创建的pod写入etcd
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	jsonstr, err := json.Marshal(pod)
	if err != nil {
		panic(err)
	}
	st.Put(constant.EtcdPodPrefix+pod.Config.Metadata.Namespace+"/"+pod.Config.Metadata.Name, jsonstr)

	c.JSON(http.StatusOK, nil)
}

// func StopPod(c *gin.Context) {
// 	var podConfig protocol.PodConfig
// 	c.BindJSON(&podConfig)
// 	var pod protocol.Pod
// 	podjson, err := json.Marshal(podConfig)
// 	if err != nil {
// 		panic(err)
// 	}
// 	json.Unmarshal(podjson, &pod.Config)
// 	msg, _ := json.Marshal(pod.Config)
// 	nodeName := GetPodNode(pod.Config)
// 	message.Publish(message.KubeletStopPodQueue+"/"+nodeName, msg)

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "stop pod: " + pod.Config.Metadata.Namespace + "/" + pod.Config.Metadata.Name,
// 	})
// }

func DeletePod(c *gin.Context) {
	var pod protocol.Pod
	c.BindJSON(&pod.Config)
	if pod.Config.Metadata.Namespace == "" {
		pod.Config.Metadata.Namespace = "default"
	}
	msg, _ := json.Marshal(pod.Config)
	nodeName := GetPodNode(pod.Config)
	message.Publish(message.KubeletDeletePodQueue+"/"+nodeName, msg)

	// 将删除pod写入etcd
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	st.Del(constant.EtcdPodPrefix + pod.Config.Metadata.Namespace + "/" + pod.Config.Metadata.Name)

	c.JSON(http.StatusOK, "delete pod: "+pod.Config.Metadata.Namespace+"/"+pod.Config.Metadata.Name)
}

func GetOnePod(c *gin.Context) {
	var rsMeta protocol.MetadataType
	c.BindJSON(&rsMeta)
	if rsMeta.Namespace == "" {
		rsMeta.Namespace = "default"
	}
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	reply, err := st.Get(constant.EtcdPodPrefix + rsMeta.Namespace + "/" + rsMeta.Name)
	if err != nil {
		logger.KError("Get One Pod error: %s", err)
		c.JSON(http.StatusBadRequest, "Get One Pod error")
		return
	}

	var pd protocol.Pod
	if len(reply.Value) == 0 {
		// 一个空的reply，返回一个空体方便解析
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	fmt.Printf("GetOneReplicaset: %s\n", string(reply.Value))
	err = json.Unmarshal(reply.Value, &pd)
	if err != nil {
		logger.KError("Parse One Replicaset error: %s", err)
		c.JSON(http.StatusBadRequest, "Parse One Replicaset error")
		return
	}

	c.JSON(http.StatusOK, pd)
}
