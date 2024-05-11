package protocol

type ReplicasetSpecType struct {
	Replicas int               `yaml:"replicas" json:"replicas"`
	Selector map[string]string `yaml:"selector" json:"selector"`
	Template PodSpecType       `yaml:"template" json:"template"`
}

type ReplicasetConfig struct {
	ApiVersion string             `yaml:"apiVersion" json:"apiVersion"`
	Kind       string             `yaml:"kind" json:"kind"`
	Metadata   MetadataType       `yaml:"metadata" json:"metadata"`
	Spec       ReplicasetSpecType `yaml:"spec" json:"spec"`
}

type ReplicasetType struct {
	Config ReplicasetConfig `yaml:"config" json:"config"`
}
