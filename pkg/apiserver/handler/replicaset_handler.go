package handler

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateReplicaset(c *gin.Context) {
	var rs protocol.ReplicasetType
	c.BindJSON(&rs)
	// data, _ := yaml.Marshal(rs)
	// fmt.Printf("CreateReplicaset: %s\n", string(data))
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

// func GetReplicaset(c *gin.Context) {
// 	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer st.Close()

// 	jsonstr, err := json.Marshal(rs)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = st.Put(constant.EtcdReplicasetPrefix+rs.Config.Metadata.Namespace+"/"+rs.Config.Metadata.Name, jsonstr)
// 	if err != nil {
// 		panic(err)
// 	}

// 	message.Publish(message.CreateReplicasetQueueName, jsonstr)

// 	c.JSON(http.StatusOK, rs)
// }
