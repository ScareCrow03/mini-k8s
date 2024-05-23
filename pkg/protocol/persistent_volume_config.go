package protocol

type MetadataNameType struct {
	Name        string            `yaml:"name" json:"name"`
	Labels      map[string]string `yaml:"labels" json:"labels"`
	Annotations map[string]string `yaml:"annotations" json:"annotations"`
}

type PersistentVolumeCapacityType struct {
	Storage    string  `yaml:"storage" json:"storage"`
	StorageNum float64 `yaml:"storageNum" json:"storageNum"`
}

type NfsType struct {
	Path   string `yaml:"path" json:"path"`
	Server string `yaml:"server" json:"server"`
}

type PersistentVolumeSpecType struct {
	Capacity PersistentVolumeCapacityType `yaml:"capacity" json:"capacity"`
	// VolumeMode                    string                       `yaml:"volumeMode" json:"volumeMode"`
	AccessModes                   string `yaml:"accessModes" json:"accessModes"` // 默认ReadWriteMany
	PersistentVolumeReclaimPolicy string `yaml:"persistentVolumeReclaimPolicy" json:"persistentVolumeReclaimPolicy"`
	StorageClassName              string `yaml:"storageClassName" json:"storageClassName"`
	// MountOptions                  []string
	Nfs NfsType `yaml:"nfs" json:"nfs"`
}

type PersistentVolume struct {
	ApiVersion string                   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                   `yaml:"kind" json:"kind"`
	Metadata   MetadataNameType         `yaml:"metadata" json:"metadata"`
	Spec       PersistentVolumeSpecType `yaml:"spec" json:"spec"`
}
