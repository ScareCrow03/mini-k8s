package protocol

import (
	"time"

	"github.com/docker/docker/api/types"
)

type MetadataType struct {
	Name        string            `yaml:"name" json:"name"`
	Namespace   string            `yaml:"namespace" json:"namespace"`
	Labels      map[string]string `yaml:"labels" json:"labels"`
	Annotations map[string]string `yaml:"annotations" json:"annotations"`
	UID         string            `yaml:"uid" json:"uid"` // 这个字段是在创建pod时由k8s自动生成的，不需要在yaml文件中指定，但是有时需要被解析出来
}

type PodSpecType struct {
	RestartPolicy string            `yaml:"restartPolicy" json:"restartPolicy"`
	Containers    []ContainerConfig `yaml:"containers" json:"containers"`
	NodeSelector  map[string]string `yaml:"nodeSelector" json:"nodeSelector"`
	Volumes       []VolumeType      `yaml:"volumes" json:"volumes"`
}

type VolumeType struct { // 默认只支持hostPath类型的volume
	Name     string       `yaml:"name" json:"name"`
	HostPath HostPathType `yaml:"hostPath" json:"hostPath"`
}

type HostPathType struct {
	Path string `yaml:"path" json:"path"`
	Type string `yaml:"type" json:"type"`
}

type PodConfig struct {
	ApiVersion string       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Metadata   MetadataType `yaml:"metadata" json:"metadata"`
	Spec       PodSpecType  `yaml:"spec" json:"spec"`
}

type PodStatus struct {
	Phase           string                 `yaml:"phase" json:"phase"` // 可以是以下几个枚举值之一：Pending、Running、Succeeded、Failed、Unknown
	Runtime         time.Duration          `yaml:"runtime" json:"runtime"`
	UpdateTime      time.Time              `yaml:"updatetime" json:"updatetime"`
	IP              string                 `yaml:"IP" json:"IP"`
	ContainerStatus []types.ContainerState `yaml:"containerStatus" json:"containerStatus"`
	NodeName        string                 `yaml:"nodeName" json:"nodeName"`
}

type Pod struct {
	Config PodConfig `yaml:"config" json:"config"`
	Status PodStatus `yaml:"status" json:"status"`
}

//func (podConfig *PodConfig) YAMLToPodConfig(path string) error {
//	file, err := os.ReadFile(path)
//	if err != nil {
//		logger.KError("read pod yaml failed")
//		return err
//	}
//
//	err = yaml.Unmarshal(file, podConfig)
//	if err != nil {
//		logger.KError("pod yaml unmarshal failed")
//		fmt.Println(err)
//		return err
//	}
//	return nil
//}
