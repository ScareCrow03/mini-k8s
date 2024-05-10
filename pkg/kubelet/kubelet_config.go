package kubelet

import (
	"mini-k8s/pkg/protocol"
	"time"
)

type KubeletConfig struct {
	ApiServerAddress string `yaml:"apiServerAddress" json:"apiServerAddress"`
	Name             string `yaml:"name" json:"name"`
	Roles            string `yaml:"roles" json:"roles"`
	Version          string `yaml:"version" json:"version"`
}

type Kubelet struct {
	Config    KubeletConfig  `yaml:"config" json:"config"`
	Status    string         `yaml:"status" json:"status"`
	StartTime time.Time      `yaml:"startTime" json:"startTime"`
	Runtime   time.Duration  `yaml:"runtime" json:"runtime"`
	Pods      []protocol.Pod `yaml:"pods" json:"pods"`
}
