package handler

import (
	"encoding/json"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 要求传一个CRType过来，允许自己的Spec字段
func CreateCR(c *gin.Context) {
	var cr protocol.CRType
	c.BindJSON(&cr)
	cr.Metadata.UID = "mini-k8s-cr-" + uid.NewUid()

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		logger.KError("etcd.NewEtcdStore error: %s", err)
	}
	defer st.Close()

	jsonstr, err := json.Marshal(cr)
	if err != nil {
		logger.KError("json.Marshal error: %s", err)
	}

	key := constant.EtcdCRPrefix + strings.ToLower(cr.Kind) + "/" + cr.Metadata.Namespace + "/" + cr.Metadata.Name
	err = st.Put(key, jsonstr)
	if err != nil {
		logger.KError("etcd.Put error: %s", err)
	}

	c.JSON(http.StatusOK, cr)
}

func DeleteCR(c *gin.Context) {
	var cr protocol.CRType
	c.BindJSON(&cr)

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		logger.KError("etcd.NewEtcdStore error: %s", err)
	}
	defer st.Close()

	key := constant.EtcdCRPrefix + strings.ToLower(cr.Kind) + "/" + cr.Metadata.Namespace + "/" + cr.Metadata.Name
	err = st.Del(key)
	if err != nil {
		logger.KError("etcd.Del error: %s", err)
	}

	c.JSON(http.StatusOK, nil)
}

// 给定crType类型串，查看该类型的CR，并返回一个切片；默认所有crType转换成小写
func GetAllCRByType(crType string) []protocol.CRType {
	crType = strings.ToLower(crType)
	// 做一些简写的转换
	if strings.Contains(crType, "ping") {
		crType = "pingsource"
	}
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		logger.KError("etcd.NewEtcdStore error: %s", err)
		return []protocol.CRType{}
	}
	defer st.Close()

	prefix := constant.EtcdCRPrefix + crType + "/"
	reply, err := st.GetWithPrefix(prefix)
	if err != nil {
		logger.KError("etcd.GetWithPrefix error: %s", err)
		return []protocol.CRType{}
	}

	var crs []protocol.CRType
	for _, r := range reply {
		var cr protocol.CRType
		err = json.Unmarshal(r.Value, &cr)
		if err != nil {
			logger.KError("json.Unmarshal error: %s", err)
			return []protocol.CRType{}
		}
		crs = append(crs, cr)
	}

	return crs
}
