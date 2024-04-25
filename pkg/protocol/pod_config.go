package protocol

import (
	"fmt"
	"mini-k8s/pkg/logger"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"gopkg.in/yaml.v3"
)

type MetadataType struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
	UID         string            `yaml:"uid"` // 这个字段是在创建pod时由k8s自动生成的，不需要在yaml文件中指定，但是有时需要被解析出来
}

type PodSpecType struct {
	RestartPolicy string            `yaml:"restartPolicy"`
	Containers    []ContainerConfig `yaml:"containers"`
	NodeSelector  map[string]string `yaml:"nodeSelector"`
	Volumes       []VolumeType      `yaml:"volumes"`
}

type VolumeType struct {
	Name     string       `yaml:"name"`
	HostPath HostPathType `yaml:"hostPath"`
}

type HostPathType struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"`
}

type PodConfig struct {
	ApiVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   MetadataType `yaml:"metadata"`
	Spec       PodSpecType  `yaml:"spec"`
}

type PodStatus struct {
	Phase           string // 可以是以下几个枚举值之一：Pending、Running、Succeeded、Failed、Unknown
	Runtime         time.Duration
	UpdateTime      time.Time
	IP              string
	ContainerStatus []types.ContainerState
}

type Pod struct {
	Config PodConfig
	Status PodStatus
}

func (podConfig *PodConfig) YAMLToPodConfig(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		logger.KError("read pod yaml failed")
		return err
	}

	err = yaml.Unmarshal(file, podConfig)
	if err != nil {
		logger.KError("pod yaml unmarshal failed")
		fmt.Print(err, "\n")
		return err
	}
	return nil
}
