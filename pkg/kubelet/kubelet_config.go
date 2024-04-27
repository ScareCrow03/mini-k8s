package kubelet

import (
	"mini-k8s/pkg/protocol"
	"time"
)

type KubeletConfig struct {
	ApiServerAddress string `yaml:"apiServerAddress"`
	Name             string `yaml:"name"`
	Roles            string `yaml:"roles"`
	Version          string `yaml:"version"`
}

type Kubelet struct {
	Config    KubeletConfig
	Status    string
	StartTime time.Time
	Age       time.Duration
	//PodService *rtm.RemoteRuntimeService
	Pods []protocol.Pod
}
