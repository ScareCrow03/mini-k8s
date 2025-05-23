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
	if cr.Metadata.Namespace == "" {
		cr.Metadata.Namespace = "default"
	}

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

	if cr.Metadata.Namespace == "" {
		cr.Metadata.Namespace = "default"
	}

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
	protocol.DoCRKindStrExpands(&crType)
	// 做一些简写的转换

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

func GetOneCR(c *gin.Context) {
	// 因为CR的Type字段在外面，这里还不能只传一个MetadataType，要传递Kind，Namespace，Name以唯一确定它
	var cr protocol.CRType
	c.BindJSON(&cr)

	if cr.Metadata.Namespace == "" {
		cr.Metadata.Namespace = "default"
	}
	protocol.DoCRKindStrExpands(&cr.Kind)

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		logger.KError("etcd.NewEtcdStore error: %s", err)
	}

	defer st.Close()

	key := constant.EtcdCRPrefix + strings.ToLower(cr.Kind) + "/" + cr.Metadata.Namespace + "/" + cr.Metadata.Name
	reply, err := st.Get(key)

	if err != nil {
		logger.KError("etcd.Get error: %s", err)
		c.JSON(http.StatusBadRequest, "Get One CR error")
		return
	}

	var crInEtcd protocol.CRType
	if len(reply.Value) == 0 { // 没有找到的情况，返回一个空体告知（没有Name字段）
		c.JSON(http.StatusOK, crInEtcd)
		return
	}
	// 还是要解析成泛型的对象，然后用gin的JSON方法返回
	err = json.Unmarshal(reply.Value, &crInEtcd)
	if err != nil {
		logger.KError("json.Unmarshal error: %s", err)
		c.JSON(http.StatusBadRequest, "Parse One CR error")
		return
	}

	c.JSON(http.StatusOK, crInEtcd)

}
