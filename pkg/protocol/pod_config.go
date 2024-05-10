package protocol

import (
	"fmt"
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
	NodeName   string       `yaml:"nodeName" json:"nodeName"`
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

// 给定一些Pods与希望它们具有的labels selector，如果Pod的元数据labels包含这个selector指定的所有匹配，则挑选成功
func SelectPodsByLabels(selector map[string]string, pods []*Pod) []*Pod {
	var selectedPods []*Pod
	for _, pod := range pods {
		if IsSelectorMatchOnePod(selector, pod) {
			selectedPods = append(selectedPods, pod)
		}
	}
	return selectedPods
}

// 单独检查一个Pod是否满足给定的Selector，上面这个函数只是一个数组的封装
func IsSelectorMatchOnePod(selector map[string]string, pod *Pod) bool {
	fmt.Printf("Pod labels: %s\n", pod.Config.Metadata.Labels)
	fmt.Printf("Selector: %s\n", selector)
	for key, val := range selector {
		if pod.Config.Metadata.Labels[key] != val {
			fmt.Printf("unmatch key: %s, val: %s, expected %s\n", key, pod.Config.Metadata.Labels[key], val)
			return false
		}
	}
	fmt.Printf("Matched!\n")
	return true
}
