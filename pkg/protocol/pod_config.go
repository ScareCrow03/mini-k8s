package protocol

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/logger"
	"os"
	"time"
)

type PodConfig struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name        string            `yaml:"name"`
		Namespace   string            `yaml:"namespace"`
		Labels      map[string]string `yaml:"labels"`
		Annotations map[string]string `yaml:"annotations"`
	} `yaml:"metadata"`
	Spec struct {
		RestartPolicy string            `yaml:"restartPolicy"`
		Containers    []ContainerConfig `yaml:"containers"`
		NodeSelector  map[string]string `yaml:"nodeSelector"`
		Volumes       []struct {
			Name     string `yaml:"name"`
			HostPath struct {
				Path string `yaml:"path"`
				Type string `yaml:"type"`
			} `yaml:"hostPath"`
		} `yaml:"volumes"`
	}
}

type PodStatus struct {
	Phase      string
	Runtime    time.Duration
	UpdateTime time.Time
	IP         string
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
