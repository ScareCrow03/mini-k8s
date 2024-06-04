package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/protocol"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 因为kubelet定期向api-server更新pod状态，所以直接从etcd中取出并返回即可
// 获取所有pod并直接返回
func GetObjectByType(c *gin.Context) {
	// test_service := rtm.NewRemoteRuntimeService(time.Minute)
	var objectType string
	c.BindJSON(&objectType)

	switch objectType {
	case "pod":
		c.JSON(http.StatusOK, GetAllPods())
	case "service":
		c.JSON(http.StatusOK, GetAllServices())
	case "dns":
		c.JSON(http.StatusOK, GetAllDns())
	case "function":
		c.JSON(http.StatusOK, GetAllFunctions())
	case "replicaset":
		c.JSON(http.StatusOK, GetAllReplicasets())
	case "hpa":
		c.JSON(http.StatusOK, GetAllHPAs())
	case "node": // 仅有kubelet的静态信息
		c.JSON(http.StatusOK, GetAllNodes())
	case "job":
		c.JSON(http.StatusOK, GetAllJobs())
	case "pv":
		c.JSON(http.StatusOK, GetAllPVs())
	case "pvc":
		c.JSON(http.StatusOK, GetAllPVCs())
	default:
		c.JSON(http.StatusOK, GetAllCRByType(objectType))
	}

}

func GetAllDns() []protocol.Dns {
	fmt.Println("get dns in etcd")
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.GetWithPrefix(constant.EtcdDnsPrefix)
	if err != nil {
		panic(err)
	}
	var dnss []protocol.Dns
	for _, r := range reply {
		var d protocol.Dns
		err = json.Unmarshal(r.Value, &d)
		if err != nil {
			panic(err)
		}
		dnss = append(dnss, d)
	}
	return dnss
}

func GetAllFunctions() []protocol.Function {
	fmt.Println("get functions in etcd")
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.GetWithPrefix(constant.EtcdFunctionPrefix)
	if err != nil {
		panic(err)
	}
	var functions []protocol.Function
	for _, r := range reply {
		var f protocol.Function
		err = json.Unmarshal(r.Value, &f)
		if err != nil {
			panic(err)
		}
		functions = append(functions, f)
	}
	return functions
}

func GetAllJobs() []protocol.Job {
	fmt.Println("get jobs in etcd")
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.GetWithPrefix(constant.EtcdJobPrefix)
	if err != nil {
		panic(err)
	}
	var jobs []protocol.Job
	for _, r := range reply {
		var j protocol.Job
		err = json.Unmarshal(r.Value, &j)
		if err != nil {
			panic(err)
		}
		jobs = append(jobs, j)
	}
	return jobs
}

func GetAllPVs() []protocol.PersistentVolume {
	fmt.Println("get pv in etcd")
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.GetWithPrefix(constant.EtcdPersistentVolumePrefix)
	if err != nil {
		panic(err)
	}
	var pvs []protocol.PersistentVolume
	for _, r := range reply {
		var p protocol.PersistentVolume
		err = json.Unmarshal(r.Value, &p)
		if err != nil {
			panic(err)
		}
		pvs = append(pvs, p)
	}
	return pvs
}

func GetAllPVCs() []protocol.PersistentVolumeClaim {
	fmt.Println("get pvc in etcd")
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.GetWithPrefix(constant.EtcdPersistentVolumeClaimPrefix)
	if err != nil {
		panic(err)
	}
	var pvcs []protocol.PersistentVolumeClaim
	for _, r := range reply {
		var p protocol.PersistentVolumeClaim
		err = json.Unmarshal(r.Value, &p)
		if err != nil {
			panic(err)
		}
		pvcs = append(pvcs, p)
	}
	return pvcs
}
