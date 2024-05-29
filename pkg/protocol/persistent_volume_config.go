package protocol

import "strconv"

type PVCapacity struct {
	Storage    string `yaml:"storage" json:"storage"`
	StorageNum int    `yaml:"storageNum" json:"storageNum"`
}

type NfsType struct {
	Path   string `yaml:"path" json:"path"`
	Server string `yaml:"server" json:"server"`
}

type PVSpec struct {
	Capacity PVCapacity `yaml:"capacity" json:"capacity"`
	// VolumeMode                    string                       `yaml:"volumeMode" json:"volumeMode"`
	AccessModes                   string `yaml:"accessModes" json:"accessModes"` // 默认ReadWriteMany
	PersistentVolumeReclaimPolicy string `yaml:"persistentVolumeReclaimPolicy" json:"persistentVolumeReclaimPolicy"`
	// StorageClassName              string `yaml:"storageClassName" json:"storageClassName"`
	// MountOptions                  []string
	Nfs NfsType `yaml:"nfs" json:"nfs"`
}

type PersistentVolume struct {
	ApiVersion  string       `yaml:"apiVersion" json:"apiVersion"`
	Kind        string       `yaml:"kind" json:"kind"`
	Metadata    MetadataType `yaml:"metadata" json:"metadata"`
	Spec        PVSpec       `yaml:"spec" json:"spec"`
	LastStorage int          `yaml:"lastStorage" json:"lastStorage"`
}

func (pv *PersistentVolume) init() {
	// TODO: 在nfs共享空间为PV分配空间
	str := pv.Spec.Capacity.Storage[:len(pv.Spec.Capacity.Storage)-2]
	num, _ := strconv.ParseFloat(str, 64)
	if str == "Gi" {
		num = num * 1024
	}
	pv.Spec.Capacity.StorageNum = int(num)
	pv.LastStorage = int(num)
}

func (pv *PersistentVolume) matchPVC(pvc PersistentVolumeClaim) bool {
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
func (pvc *PersistentVolumeClaim) matchPV(pv PersistentVolume) bool {
	return pv.matchPVC(*pvc)
}
