package protocol

import "mini-k8s/pkg/constant"

type Function struct {
	ApiVersion string       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Metadata   MetadataType `yaml:"metadata" json:"metadata"`
	Spec       FunctionSpec `yaml:"spec" json:"spec"`
}

type FunctionSpec struct {
	UserUploadFile []byte `yaml:"userUploadFile" json:"userUploadFile"`
	UserUploadPath string `yaml:"userUploadFilePath" json:"userUploadFilePath"`
}

func GetOneReplicaConfigFromFunction(f Function) ReplicasetConfig {
	var replica ReplicasetConfig
	replica.ApiVersion = "v1"
	replica.Kind = "Replicaset"
	replica.Metadata.Namespace = f.Metadata.Namespace
	replica.Metadata.Name = f.Metadata.Name
	replica.Spec.Replicas = 1
	replica.Spec.Selector.MatchLabels = make(map[string]string)
	replica.Spec.Template.Metadata.Labels = make(map[string]string)
	replica.Spec.Selector.MatchLabels["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name
	replica.Spec.Template.Metadata.Labels["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name
	replica.Spec.Template.Spec.Containers = make([]ContainerConfig, 1)
	replica.Spec.Template.Spec.Containers[0].Name = f.Metadata.Name
	replica.Spec.Template.Spec.Containers[0].Image = constant.BaseImage + "/" + f.Metadata.Namespace + "/" + f.Metadata.Name + ":latest"
	replica.Spec.Template.Spec.Containers[0].Ports = make([]CtrPortBindingType, 1)
	replica.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = 10000
	return replica
}

func GetServiceConfigFromFunction(f Function) ServiceConfig {
	var service ServiceConfig
	service.ApiVersion = "v1"
	service.Kind = "Service"
	service.Metadata.Namespace = f.Metadata.Namespace
	service.Metadata.Name = f.Metadata.Name
	service.Spec.Ports = make([]ServicePort, 1)
	service.Spec.Selector = make(map[string]string)
	service.Spec.Ports[0].Port = 10000
	service.Spec.Ports[0].TargetPort = 10000

	service.Spec.Selector["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name

	// 添加function相关的信息到service自己的labels上
	service.Metadata.Labels = make(map[string]string)
	service.Metadata.Labels["type"] = "function"
	service.Metadata.Labels["functionName"] = f.Metadata.Name
	service.Metadata.Labels["functionNamespace"] = f.Metadata.Namespace
	return service
}
