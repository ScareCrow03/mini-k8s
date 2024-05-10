package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"net/http"
	"strconv"

	"math/rand"

	"github.com/gin-gonic/gin"
)

func GetClusterIP() string {
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	var cip string
	for {
		cip = "222.111." + strconv.Itoa(rand.Int()%256) + "." + strconv.Itoa(rand.Int()%256)
		reply, _ := st.Get(constant.EtcdServicePrefix + "clusterIP" + "/" + cip)
		if len(reply.Value) == 0 {
			break
		}
	}

	jsonstr, err := json.Marshal(cip)
	if err != nil {
		panic(err)
	}

	st.Put(constant.EtcdServicePrefix+"clusterIP"+"/"+cip, jsonstr)

	return cip
}

func DelClusterIP(cip string) string {
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	err = st.Del(constant.EtcdServicePrefix + "clusterIP" + "/" + cip)
	if err != nil {
		panic(err)
	}

	return cip
}

func CreateService(c *gin.Context) {
	var svc protocol.ServiceType
	c.BindJSON(&svc)

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	svc.Config.Spec.ClusterIP = GetClusterIP()
	fmt.Println("CreateService cluster IP: ", svc.Config.Spec.ClusterIP)

	jsonstr, err := json.Marshal(svc)
	if err != nil {
		panic(err)
	}
	err = st.Put(constant.EtcdServicePrefix+svc.Config.Metadata.Namespace+"/"+svc.Config.Metadata.Name, jsonstr)
	if err != nil {
		panic(err)
	}

	message.Publish(message.CreateServiceQueueName, jsonstr)

	c.JSON(http.StatusOK, svc)
}

func DeleteService(c *gin.Context) {
	var svc protocol.ServiceType
	c.BindJSON(&svc)

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	jsonstr, err := json.Marshal(svc)
	if err != nil {
		panic(err)
	}

	DelClusterIP(svc.Config.Spec.ClusterIP)
	err = st.Del(constant.EtcdServicePrefix + svc.Config.Metadata.Namespace + "/" + svc.Config.Metadata.Name)
	if err != nil {
		panic(err)
	}

	message.Publish(message.DeleteServiceQueueName, jsonstr)

	c.JSON(http.StatusOK, nil)
}
