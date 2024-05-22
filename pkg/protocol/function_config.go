package protocol

type Function struct {
	ApiVersion string       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Metadata   MetadataType `yaml:"metadata" json:"metadata"`
	Spec       FunctionSpec `yaml:"spec" json:"spec"`
}

type FunctionSpec struct {
	UserUploadFile []byte `yaml:"userUploadFile" json:"userUploadFile"`
	UserUploadPath string `yaml:"userUploadPath" json:"userUploadPath"`
}
