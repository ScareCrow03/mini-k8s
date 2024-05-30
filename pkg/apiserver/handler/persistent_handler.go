package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/protocol"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func CreatePV(c *gin.Context) {
	var pv protocol.PersistentVolume
	c.BindJSON(&pv)
	pv.Spec.Capacity.StorageNum = protocol.Storage2Num(pv.Spec.Capacity.Storage)
	pv.Status = "Available"

	// PV没有namespace
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.Get(constant.EtcdPersistentVolumePrefix + pv.Metadata.Name)
	if err != nil {
		panic(err)
	}
	if len(reply.Value) > 0 { // 存在重复pv，阻止
		fmt.Println("Create PV from file failed: same PV name")
		c.JSON(http.StatusOK, "Create PV from file failed: same PV name")
		return
	}

	os.RemoveAll(constant.PersistentDir + pv.Metadata.Name)
	err = os.Mkdir(constant.PersistentDir+pv.Metadata.Name, os.ModePerm)
	if err != nil {
		panic(err)
	}

	jsonstr, err := json.Marshal(pv)
	if err != nil {
		panic(err)
	}
	st.Put(constant.EtcdPersistentVolumePrefix+pv.Metadata.Name, jsonstr)

	c.JSON(http.StatusOK, "create PV from file: "+pv.Metadata.Name)
}

func DeletePV(c *gin.Context) {
	var pv protocol.PersistentVolume
	c.BindJSON(&pv)
	pv.Spec.Capacity.StorageNum = protocol.Storage2Num(pv.Spec.Capacity.Storage)

	// PV没有namespace
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	// 取出PV
	reply, err := st.Get(constant.EtcdPersistentVolumePrefix + pv.Metadata.Name)
	if err != nil {
		panic(err)
	}
	if len(reply.Value) == 0 {
		c.JSON(http.StatusOK, "delete PV from file failed, PV not exists: "+pv.Metadata.Name)
		return
	}
	err = st.Del(constant.EtcdPersistentVolumePrefix + pv.Metadata.Name)
	if err != nil {
		panic(err)
	}

	// 删除etcd中该PV对应的PVC
	reply1, err := st.GetWithPrefix(constant.EtcdPersistentVolumeClaimPrefix)
	if err != nil {
		panic(err)
	}
	var pvc protocol.PersistentVolumeClaim
	for _, r := range reply1 {
		err = json.Unmarshal(r.Value, &pvc)
		if err != nil {
			panic(err)
		}
		if pvc.PVName == pv.Metadata.Name {
			st.Del(constant.EtcdPersistentVolumeClaimPrefix + pvc.Metadata.Namespace + "." + pvc.Metadata.Name)
		}
	}

	os.RemoveAll(constant.PersistentDir + pv.Metadata.Name) // 回收

	c.JSON(http.StatusOK, "delete PV from file: "+pv.Metadata.Name)
}

func CreatePVC(c *gin.Context) {
	var pvc protocol.PersistentVolumeClaim
	c.BindJSON(&pvc)
	pvc.Spec.Resources.Requests.StorageNum = protocol.Storage2Num(pvc.Spec.Resources.Requests.Storage)
	if pvc.Metadata.Namespace == "" {
		pvc.Metadata.Namespace = "default"
	}

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.Get(constant.EtcdPersistentVolumeClaimPrefix + pvc.Metadata.Namespace + "/" + pvc.Metadata.Name)
	if err != nil {
		panic(err)
	}
	if len(reply.Value) > 0 { // 存在重复pvc，阻止
		fmt.Println("Create PVC from file failed: same PVC name")
		c.JSON(http.StatusOK, "Create PVC from file failed: same PVC name")
		return
	}

	// 取出所有PV，找一个匹配的
	reply1, err := st.GetWithPrefix(constant.EtcdPersistentVolumePrefix)
	if err != nil {
		panic(err)
	}
	var pv protocol.PersistentVolume
	var found bool = false
	for _, r := range reply1 {
		err = json.Unmarshal(r.Value, &pv)
		if err != nil {
			panic(err)
		}
		if protocol.PVmatchPVC(pv, pvc) {
			found = true
			break
		}
	}

	if !found {
		// TODO: 自动创建对应的PV
		pv.ApiVersion = pvc.ApiVersion
		pv.Kind = "persistentVolume"
		pv.Metadata.Name = pvc.Metadata.Name + "_autoCreatePV"
		pv.Metadata.Labels = make(map[string]string)
		for key, value := range pvc.Spec.Selector.MatchLabels {
			pv.Metadata.Labels[key] = value
		}
		pv.Spec.Capacity.Storage = pvc.Spec.Resources.Requests.Storage
		pv.Spec.Capacity.StorageNum = pvc.Spec.Resources.Requests.StorageNum
		pv.Spec.AccessModes = pvc.Spec.AccessModes
		pv.Spec.PersistentVolumeReclaimPolicy = "Delete" // 删除PVC时，也删除PV
		pv.Status = "Available"

		os.RemoveAll(constant.PersistentDir + pv.Metadata.Name)
		err = os.Mkdir(constant.PersistentDir+pv.Metadata.Name, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	os.RemoveAll(constant.PersistentDir + pv.Metadata.Name + "/" + pvc.Metadata.Namespace + "." + pvc.Metadata.Name)
	err = os.Mkdir(constant.PersistentDir+pv.Metadata.Name+"/"+pvc.Metadata.Namespace+"."+pvc.Metadata.Name, os.ModePerm)
	if err != nil {
		panic(err)
	}

	pv.Status = "Bound"
	jsonstr, err := json.Marshal(pv)
	if err != nil {
		panic(err)
	}
	st.Put(constant.EtcdPersistentVolumePrefix+pv.Metadata.Name, jsonstr)

	pvc.PVName = pv.Metadata.Name
	jsonstr, err = json.Marshal(pvc)
	if err != nil {
		panic(err)
	}
	st.Put(constant.EtcdPersistentVolumeClaimPrefix+pvc.Metadata.Namespace+"."+pvc.Metadata.Name, jsonstr)

	fmt.Println("create PVC from file: " + pvc.Metadata.Namespace + "." + pvc.Metadata.Name + ", in PV " + pvc.PVName)

	c.JSON(http.StatusOK, "create PVC from file: "+pvc.Metadata.Namespace+"."+pvc.Metadata.Name+", in PV "+pvc.PVName)
}

func DeletePVC(c *gin.Context) {
	var pvc protocol.PersistentVolumeClaim
	c.BindJSON(&pvc)
	pvc.Spec.Resources.Requests.StorageNum = protocol.Storage2Num(pvc.Spec.Resources.Requests.Storage)
	if pvc.Metadata.Namespace == "" {
		pvc.Metadata.Namespace = "default"
	}

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	// 取出PVC
	reply, err := st.Get(constant.EtcdPersistentVolumeClaimPrefix + pvc.Metadata.Namespace + "." + pvc.Metadata.Name)
	if err != nil {
		panic(err)
	}
	if len(reply.Value) == 0 {
		c.JSON(http.StatusOK, "delete PVC from file failed, PVC not exists: "+pvc.Metadata.Name)
		return
	}
	err = json.Unmarshal(reply.Value, &pvc)
	if err != nil {
		panic(err)
	}

	// 取出PV
	var pv protocol.PersistentVolume
	reply, err = st.Get(constant.EtcdPersistentVolumePrefix + pvc.PVName)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(reply.Value, &pv)
	if err != nil {
		panic(err)
	}

	// 释放PVC占据的PV
	if pv.Spec.PersistentVolumeReclaimPolicy == "Retain" { // 保留PV
		os.RemoveAll(constant.PersistentDir + pv.Metadata.Name + "/" + pvc.Metadata.Namespace + "." + pvc.Metadata.Name) // 仅删除PVC
		pv.Status = "Available"
		jsonstr, err := json.Marshal(pv)
		if err != nil {
			panic(err)
		}
		st.Put(constant.EtcdPersistentVolumePrefix+pv.Metadata.Name, jsonstr)
	} else { // 删除PV
		os.RemoveAll(constant.PersistentDir + pv.Metadata.Name)
		st.Del(constant.EtcdPersistentVolumePrefix + pv.Metadata.Name)
	}

	st.Del(constant.EtcdPersistentVolumeClaimPrefix + pvc.Metadata.Namespace + "." + pvc.Metadata.Name)

	c.JSON(http.StatusOK, "delete PVC from file: "+pvc.Metadata.Name)
}

func GetPVC(c *gin.Context) {
	var pvc protocol.PersistentVolumeClaim
	c.BindJSON(&pvc)
	if pvc.Metadata.Namespace == "" {
		pvc.Metadata.Namespace = "default"
	}

	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.Get(constant.EtcdPersistentVolumeClaimPrefix + pvc.Metadata.Namespace + "." + pvc.Metadata.Name)
	if err != nil {
		panic(err)
	}
	if len(reply.Value) == 0 { // 不存在pvc，阻止
		fmt.Println("Get PVC failed: PVC not exists, ", pvc.Metadata.Namespace+"."+pvc.Metadata.Name)
		c.JSON(http.StatusOK, pvc)
		return
	}
	err = json.Unmarshal(reply.Value, &pvc)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, pvc)
}
