package protocol

type ReplicaSetType struct {
	Config ReplicaSetConfig `yaml:"config" json:"config"`
}

type ReplicaSetConfig struct {
	ApiVersion string             `yaml:"apiVersion" json:"apiVersion"`
	Kind       string             `yaml:"kind" json:"kind"`
	Metadata   MetadataType       `yaml:"metadata" json:"metadata"`
	Spec       ReplicaSetSpecType `yaml:"spec" json:"spec"`
}

type ReplicaSetSpecType struct {
	Replicas int               `yaml:"replicas" json:"replicas"`
	Selector map[string]string `yaml:"selector" json:"selector"`
	Template PodSpecType       `yaml:"template" json:"template"`
}
