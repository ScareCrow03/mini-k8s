package protocol

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
	Resources   PVCResources `yaml:"resources" json:"resources"`
	Selector    PVCSelector  `yaml:"selector" json:"selector"`
	// StorageClassName string       `yaml:"storageClassName" json:"storageClassName"`
}

type PersistentVolumeClaim struct {
	ApiVersion string       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Metadata   MetadataType `yaml:"metadata" json:"metadata"`
	Spec       PVCSpec      `yaml:"spec" json:"spec"`
	PVName     string       `yaml:"pvname" json:"pvname"`
}
