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
	NodeName   string       `yaml:"nodeName" json:"nodeName"`
}

type PodStatus struct {
	Phase           string                          `yaml:"phase" json:"phase"` // 可以是以下几个枚举值之一：Pending、Running、Succeeded、Failed、Unknown
	Runtime         time.Duration                   `yaml:"runtime" json:"runtime"`
	UpdateTime      time.Time                       `yaml:"updatetime" json:"updatetime"`
	IP              string                          `yaml:"IP" json:"IP"`
	ContainerStatus map[string]types.ContainerState `yaml:"containerStatus" json:"containerStatus"` //用ID来索引running状态字段，这个字段过于简单，只显示容器是否Running

	// 以下是自定义的状态量，用于hpa
	CtrsMetrics map[string]CtrMetricsEntry `yaml:"ctrsMetrics" json:"ctrsMetrics"`
	PodMetrics  PodMetricsEntry            `yaml:"podMetrics" json:"podMetrics"`
}

// 这是一个Pod级别的监控指标，包括CPU和内存的使用率；在ReplicaSet级别的资源使用指标，简单实现为各个Pod的指标平均值
type PodMetricsEntry struct {
	ID               string  `yaml:"id" json:"id"`
	CPUPercentage    float64 `yaml:"cpuPercentage" json:"cpuPercentage"`
	MemoryPercentage float64 `yaml:"memoryPercentage" json:"memoryPercentage"`
}

type CtrMetricsEntry struct {
	ID               string  `yaml:"id" json:"id"`
	CPUPercentage    float64 `yaml:"cpuPercentage" json:"cpuPercentage"`
	Memory           float64 `yaml:"memory" json:"memory"`
	MemoryLimit      float64 `yaml:"memoryLimit" json:"memoryLimit"`
	MemoryPercentage float64 `yaml:"memoryPercentage" json:"memoryPercentage"`
}

type Pod struct {
	Config PodConfig `yaml:"config" json:"config"`
	Status PodStatus `yaml:"status" json:"status"`
}

func ParseDockerCtrStatsToMetricsEntry(statsJson *types.StatsJSON) CtrMetricsEntry {
	// 计算CPU使用率，一般是0~1的值；如果容器使用了超过1个CPU的资源，那么可能会大于1
	cpuDelta := float64(statsJson.CPUStats.CPUUsage.TotalUsage) - float64(statsJson.PreCPUStats.CPUUsage.TotalUsage)

	systemDelta := float64(statsJson.CPUStats.SystemUsage) - float64(statsJson.PreCPUStats.SystemUsage)

	onlineCPUs := float64(statsJson.CPUStats.OnlineCPUs)
	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(statsJson.CPUStats.CPUUsage.PercpuUsage))
	}

	cpuPercentage := (cpuDelta / systemDelta) * onlineCPUs

	// 计算内存使用率
	memory := float64(statsJson.MemoryStats.Usage)
	memoryLimit := float64(statsJson.MemoryStats.Limit)
	memoryPercentage := 0.0
	if memoryLimit > 0.0 {
		memoryPercentage = memory / memoryLimit
	}

	return CtrMetricsEntry{
		ID:               statsJson.ID,
		CPUPercentage:    cpuPercentage,
		Memory:           memory,
		MemoryLimit:      memoryLimit,
		MemoryPercentage: memoryPercentage,
	}
}

func CalculatePodMertrics(podId string, containerStats []CtrMetricsEntry) PodMetricsEntry {
	var totalCPUPercentage float64
	var totalMemory float64
	var totalMemoryLimit float64
	var totalMemoryPercentage float64

	for _, containerStat := range containerStats {
		totalCPUPercentage += containerStat.CPUPercentage
		totalMemory += containerStat.Memory
		totalMemoryLimit += containerStat.MemoryLimit
	}

	if totalMemoryLimit > 0.0 {
		totalMemoryPercentage = (totalMemory / totalMemoryLimit)
	}

	return PodMetricsEntry{
		ID:               podId,
		CPUPercentage:    totalCPUPercentage,
		MemoryPercentage: totalMemoryPercentage,
	}
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
	for key, val := range selector {
		if pod.Config.Metadata.Labels[key] != val {
			return false
		}
	}
	return true
}

// 以下函数为非指针版本
func SelectPodsByLabelsNoPointer(selector map[string]string, pods []Pod) []Pod {
	var selectedPods []Pod
	for _, pod := range pods {
		if IsSelectorMatchOnePodNoPointer(selector, pod) {
			selectedPods = append(selectedPods, pod)
		}
	}
	return selectedPods
}
func IsSelectorMatchOnePodNoPointer(selector map[string]string, pod Pod) bool {
	for key, val := range selector {
		if pod.Config.Metadata.Labels[key] != val {
			return false
		}
	}
	return true
}
