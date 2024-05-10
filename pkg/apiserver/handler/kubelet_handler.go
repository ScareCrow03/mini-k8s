package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	kubelet2 "mini-k8s/pkg/kubelet"
	"net/http"

	"github.com/gin-gonic/gin"
)

func KubeletRegister(c *gin.Context) {
	// TODO: register kubelet to apiserver, write into etcd
	var kubelet kubelet2.Kubelet
	c.BindJSON(&kubelet.Config)
	fmt.Println(kubelet.Config.Name)
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

	c.JSON(http.StatusOK, gin.H{
		"message": "kubelet register: " + kubelet.Config.Name,
	})
}
