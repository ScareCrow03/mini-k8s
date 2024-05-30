package protocol

import (
	"strconv"
)

type PVCapacity struct {
	Storage    string `yaml:"storage" json:"storage"`
	StorageNum int    `yaml:"storageNum" json:"storageNum"`
}

// type NfsType struct {
// 	Path   string `yaml:"path" json:"path"`
// 	Server string `yaml:"server" json:"server"`
// }

type PVSpec struct {
	Capacity PVCapacity `yaml:"capacity" json:"capacity"`
	// VolumeMode                    string                       `yaml:"volumeMode" json:"volumeMode"`
	AccessModes                   string `yaml:"accessModes" json:"accessModes"` // 默认ReadWriteMany
	PersistentVolumeReclaimPolicy string `yaml:"persistentVolumeReclaimPolicy" json:"persistentVolumeReclaimPolicy"`
	// StorageClassName              string `yaml:"storageClassName" json:"storageClassName"`
	// MountOptions                  []string
	// Nfs NfsType `yaml:"nfs" json:"nfs"`
}

type PersistentVolume struct {
	ApiVersion  string       `yaml:"apiVersion" json:"apiVersion"`
	Kind        string       `yaml:"kind" json:"kind"`
	Metadata    MetadataType `yaml:"metadata" json:"metadata"` // 注意PV不属于任何namespace
	Spec        PVSpec       `yaml:"spec" json:"spec"`
	LastStorage int          `yaml:"lastStorage" json:"lastStorage"`
}

func Storage2Num(storage string) int {
	// fmt.Println(storage)
	str := storage[:len(storage)-2]
	num, _ := strconv.ParseFloat(str, 64)
	if str == "Gi" {
		num = num * 1024
	}
	return int(num)
}

func PVmatchPVC(pv PersistentVolume, pvc PersistentVolumeClaim) bool {
	if pv.Spec.AccessModes != pvc.Spec.AccessModes {
		return false
	}
	if pv.Spec.Capacity.StorageNum < pvc.Spec.Resources.Requests.StorageNum {
		return false
	}
	for key, value := range pvc.Spec.Selector.MatchLabels {
		v, ok := pv.Metadata.Labels[key]
		if !ok || v != value {
			return false
		}
	}
	return true
}
