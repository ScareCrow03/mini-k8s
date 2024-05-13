package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func CreateHPA(c *gin.Context) {
	var hpa protocol.HPAType
	c.BindJSON(&hpa.Config)

	data, _ := yaml.Marshal(hpa)
	fmt.Printf("CreateHPA: %s\n", string(data))

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		logger.KError("etcd.NewEtcdStore error: %s", err)
	}
	defer st.Close()

	hpa.Config.Metadata.UID = "mini-k8s-hpa-" + uid.NewUid()

	jsonstr, err := yaml.Marshal(hpa)
	if err != nil {
		logger.KError("yaml.Marshal error: %s", err)
	}

	err = st.Put(constant.EtcdHPAPrefix+hpa.Config.Metadata.Namespace+"/"+hpa.Config.Metadata.Name, jsonstr)
	if err != nil {
		logger.KError("etcd.Put error: %s", err)
	}

	c.JSON(http.StatusOK, hpa)
}

func DeleteHPA(c *gin.Context) {
	var hpa protocol.HPAType
	c.BindJSON(&hpa.Config)

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		logger.KError("etcd.NewEtcdStore error: %s", err)
	}
	defer st.Close()

	err = st.Del(constant.EtcdHPAPrefix + hpa.Config.Metadata.Namespace + "/" + hpa.Config.Metadata.Name)

	if err != nil {
		logger.KError("etcd.Del error: %s", err)
	}

	c.JSON(http.StatusOK, nil)
}

func GetAllHPAs() []protocol.HPAType {
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		logger.KError("etcd.NewEtcdStore error: %s", err)
	}
	defer st.Close()

	reply, err := st.GetWithPrefix(constant.EtcdHPAPrefix)
	if err != nil {
		logger.KError("etcd.GetWithPrefix error: %s", err)
	}

	var hpas []protocol.HPAType
	for _, r := range reply {
		var hpa protocol.HPAType
		err = json.Unmarshal(r.Value, &hpa)
		if err != nil {
			logger.KError("yaml.Unmarshal error: %s", err)
			return []protocol.HPAType{}
		}
		hpas = append(hpas, hpa)
	}
	return hpas
}
