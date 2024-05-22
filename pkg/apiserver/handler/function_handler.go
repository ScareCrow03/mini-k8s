package handler

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateFunction(c *gin.Context) {
	var function protocol.Function
	c.BindJSON(&function)
	function.Metadata.UID = "mini-k8s-function-" + uid.NewUid()
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	defer st.Close()
	jsonstr, err := json.Marshal(function)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	st.Put(constant.EtcdFunctionPrefix+function.Metadata.Namespace+"/"+function.Metadata.Name, jsonstr)
	c.JSON(http.StatusOK, gin.H{
		"message": "create function: " + function.Metadata.Namespace + "/" + function.Metadata.Name,
	})
}
