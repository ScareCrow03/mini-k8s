package protocol

type ReplicasetSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels" json:"matchLabels"`
}

type ReplicasetSpec struct {
	Replicas int                `yaml:"replicas" json:"replicas"`
	Selector ReplicasetSelector `yaml:"selector" json:"selector"`
	Template PodConfig          `yaml:"template" json:"template"`
}

type ReplicasetConfig struct {
	ApiVersion string         `yaml:"apiVersion" json:"apiVersion"`
	Kind       string         `yaml:"kind" json:"kind"`
	Metadata   MetadataType   `yaml:"metadata" json:"metadata"`
	Spec       ReplicasetSpec `yaml:"spec" json:"spec"`
}

type ReplicasetType struct {
	Config ReplicasetConfig `yaml:"config" json:"config"`
}
