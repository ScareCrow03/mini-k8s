package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
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
	c.BindJSON(&rs)
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

	message.Publish(message.CreateReplicasetQueueName, jsonstr)

	c.JSON(http.StatusOK, rs)
}

func DeleteReplicaset(c *gin.Context) {
	var rs protocol.ReplicasetType
	c.BindJSON(&rs)

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	jsonstr, err := json.Marshal(rs)
	if err != nil {
		panic(err)
	}

	err = st.Del(constant.EtcdReplicasetPrefix + rs.Config.Metadata.Namespace + "/" + rs.Config.Metadata.Name)
	if err != nil {
		panic(err)
	}

	message.Publish(message.DeleteReplicasetQueueName, jsonstr)

	c.JSON(http.StatusOK, nil)
}

func GetReplicaset(c *gin.Context) {
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	reply, err := st.GetWithPrefix(constant.EtcdPodPrefix)
	if err != nil {
		panic(err)
	}
	var rss []protocol.ReplicasetType
	for _, r := range reply {
		var rs protocol.ReplicasetType
		err = json.Unmarshal(r.Value, &rs)
		if err != nil {
			panic(err)
		}
		rss = append(rss, rs)
	}

	c.JSON(http.StatusOK, rss)
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
	var rss []protocol.ReplicasetType
	for _, r := range reply {
		var rs protocol.ReplicasetType
		err = json.Unmarshal(r.Value, &rs)
		if err != nil {
			panic(err)
		}
		rss = append(rss, rs)
	}
	return rss
}
