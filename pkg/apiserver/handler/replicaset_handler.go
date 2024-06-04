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
	"gopkg.in/yaml.v3"
)

// 创建replicaset时，selector.matchLabels与template.metadate.labels相同
// 在创建pod时，将replicaset元数据放入pod.metadata.labels，便于查找
func CreateReplicaset(c *gin.Context) {
	var rs protocol.ReplicasetType
	c.BindJSON(&rs.Config)
	if rs.Config.Metadata.Namespace == "" {
		rs.Config.Metadata.Namespace = "default"
	}
	rs.Config.Spec.Template.ApiVersion = rs.Config.ApiVersion
	rs.Config.Spec.Template.Kind = "Pod"
	rs.Config.Spec.Template.Metadata.Labels["ReplicasetMetadata"] = rs.Config.Metadata.Namespace + "/" + rs.Config.Metadata.Name

	data, _ := yaml.Marshal(rs)
	fmt.Printf("CreateReplicaset: %s\n", string(data))

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	rs.Config.Metadata.UID = "mini-k8s-replicaset-" + uid.NewUid()

	jsonstr, err := json.Marshal(rs)
	if err != nil {
		panic(err)
	}
	err = st.Put(constant.EtcdReplicasetPrefix+rs.Config.Metadata.Namespace+"/"+rs.Config.Metadata.Name, jsonstr)
	if err != nil {
		panic(err)
	}

	// message.Publish(message.CreateReplicasetQueueName, jsonstr)

	c.JSON(http.StatusOK, rs)
}

func DeleteReplicaset(c *gin.Context) {
	var rs protocol.ReplicasetType
	c.BindJSON(&rs.Config)
	if rs.Config.Metadata.Namespace == "" {
		rs.Config.Metadata.Namespace = "default"
	}
	msg, _ := json.Marshal(rs.Config)
	message.Publish(message.DeleteReplicasetQueueName, msg)

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	err = st.Del(constant.EtcdReplicasetPrefix + rs.Config.Metadata.Namespace + "/" + rs.Config.Metadata.Name)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, nil)
}

func GetOneReplicaset(c *gin.Context) {
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

	reply, err := st.Get(constant.EtcdReplicasetPrefix + rsMeta.Namespace + "/" + rsMeta.Name)
	if err != nil {
		logger.KError("Get One Replicaset error: %s", err)
		c.JSON(http.StatusBadRequest, "Get One Replicaset error")
		return
	}
	var rs protocol.ReplicasetType
	if len(reply.Value) == 0 {
		// 一个空的reply，返回一个空体方便解析
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	fmt.Printf("GetOneReplicaset: %s\n", string(reply.Value))
	err = json.Unmarshal(reply.Value, &rs)
	if err != nil {
		logger.KError("Parse One Replicaset error: %s", err)
		c.JSON(http.StatusBadRequest, "Parse One Replicaset error")
		return
	}

	c.JSON(http.StatusOK, rs)
}

func GetAllReplicasets() []protocol.ReplicasetType {
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.GetWithPrefix(constant.EtcdReplicasetPrefix)
	if err != nil {
		panic(err)
	}
	fmt.Println("GetAllReplicasets")
	var rss []protocol.ReplicasetType
	for _, r := range reply {
		fmt.Println(string(r.Value))
		var rs protocol.ReplicasetType
		err = json.Unmarshal(r.Value, &rs)
		if err != nil {
			panic(err)
		}
		rss = append(rss, rs)
	}
	return rss
}

// 修改replica数量
func ChangeReplicasetNum(c *gin.Context) {
	var rsnew protocol.ReplicasetSimpleType
	c.BindJSON(&rsnew)
	if rsnew.Namespace == "" {
		rsnew.Namespace = "default"
	}

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	reply, err := st.Get(constant.EtcdReplicasetPrefix + rsnew.Namespace + "/" + rsnew.Name)
	if err != nil {
		panic(err)
	}
	if len(reply.Value) == 0 {
		c.JSON(http.StatusOK, "replicaset not exists: "+rsnew.Name)
		return
	}

	var rs protocol.ReplicasetType
	err = json.Unmarshal(reply.Value, &rs)
	if err != nil {
		panic(err)
	}

	rs.Config.Spec.Replicas = rsnew.Replicas

	jsonstr, err := json.Marshal(rs)
	if err != nil {
		panic(err)
	}
	err = st.Put(constant.EtcdReplicasetPrefix+rs.Config.Metadata.Namespace+"/"+rs.Config.Metadata.Name, jsonstr)
	if err != nil {
		panic(err)
	}

	// jsonstr, _ = json.Marshal("replicasetCheckNow")
	message.Publish(message.ReplicasetCheckNowQueueName, jsonstr) // 解决延迟问题

	c.JSON(http.StatusOK, nil)
}
