package protocol

import "strconv"

type PVCRequests struct {
	Storage    string `yaml:"storage" json:"storage"`
	StorageNum int    `yaml:"storageNum" json:"storageNum"`
}

type PVCResources struct {
	Requests PVCRequests `yaml:"requests" json:"requests"`
}

type PVCSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels" json:"matchLabels"`
}

type PVCSpec struct {
	AccessModes string       `yaml:"accessModes" json:"accessModes"` // 默认ReadWriteMany
	Resources   PVCResources `yaml:"resouces" json:"resouces"`
	Selector    PVCSelector  `yaml:"selector" json:"selector"`
	// StorageClassName string       `yaml:"storageClassName" json:"storageClassName"`
}

type PersistentVolumeClaim struct {
	ApiVersion string       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Metadata   MetadataType `yaml:"metadata" json:"metadata"`
	Spec       PVCSpec      `yaml:"spec" json:"spec"`
}

func (pvc *PersistentVolumeClaim) init() {
	// TODO: 在nfs共享空间为PV分配空间
	str := pvc.Spec.Resources.Requests.Storage[:len(pvc.Spec.Resources.Requests.Storage)-2]
	num, _ := strconv.ParseFloat(str, 64)
	if str == "Gi" {
		pvc.Spec.Resources.Requests.StorageNum = int(num * 1024)
	} else if str == "Mi" {
		pvc.Spec.Resources.Requests.StorageNum = int(num)
	}

}
