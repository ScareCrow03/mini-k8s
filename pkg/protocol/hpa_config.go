package protocol

import (
	"encoding/json"
	"fmt"
	"math"
)

// 必须假定cpu和memory在监控指定时，至多只出现一次！否则发生资源利用率计算的错误！
type ResourceMetric struct {
	Name        string  `yaml:"name" json:"name"`               // 资源名称，只支持cpu和memory这两个字符串
	TargetValue float64 `yaml:"targetValue" json:"targetValue"` // 目标值，一个0~1的浮点，如CPU利用率50%
}

type HPAConfig struct {
	ApiVersion string       `yaml:"apiversion" json:"apiversion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Metadata   MetadataType `yaml:"metadata" json:"metadata"`
	Spec       HPASpecType  `yaml:"spec" json:"spec"`
}

type HPASpecType struct {
	ScaleTargetRef ScaleTargetRefType `yaml:"scaleTargetRef" json:"scaleTargetRef"` // 扩容的目标
	MinReplicas    int                `yaml:"minReplicas" json:"minReplicas"`       // 扩容Pod数量的下限
	MaxReplicas    int                `yaml:"maxReplicas" json:"maxReplicas"`       // 扩容Pod数量的上限
	ScaleInterval  int                `yaml:"scaleInterval" json:"scaleInterval"`   // 扩容间隔
	Metrics        []ResourceMetric   `yaml:"metrics" json:"metrics"`               // 指标类型和目标值
}

// 被管理者的目标类型
type ScaleTargetRefType struct {
	Kind string `yaml:"kind" json:"kind"` // 目标类型，只支持ReplicaSet
	Name string `yaml:"name" json:"name"` // 目标名称，此处不需要再指定namespace，默认用metadata中的namespace；本Hpa利用的是被管理replicaSet的replicas重新指定副本数，所以必须保证该replicaSet预先存在template等静态信息！
}

// 可能不需要这个状态量
// type HPAStatus struct {
// 	CurrentReplicas int    `yaml:"currentReplicas" json:"currentReplicas"` // 当前副本数
// 	DesiredReplicas int    `yaml:"desiredReplicas" json:"desiredReplicas"` // 期望的副本数
// 	LastScaleTime   string `yaml:"lastScaleTime" json:"lastScaleTime"`     // 上次扩缩容的时间
// }

type HPAType struct {
	Config HPAConfig `yaml:"config" json:"config"`
	// Status HPAStatus `yaml:"status" json:"status"`
}

// ReplicaSet自身并不需要监测资源，而是在它之上，有一个HPA要求时，才需要监测HPA给定的指标，其他指标也都不需要
type ReplicaMetricsResult struct {
	Metrics map[string]float64
}

// 有多项监控指标时，例如replicaSet级别的cpu和memory使用百分比（就是各个Pods的使用百分比均值）；需要先计算每一个单项的disiredReplicas，然后取最大值，这样作为整个HPA的disiredReplicas
func CalculateDesiredReplicas(h *HPAType, curReplicaNum int, curMetrics ReplicaMetricsResult) int {
	var maxDesiredReplicas int = 0

	data, _ := json.Marshal(curMetrics)
	fmt.Printf("CalculateDesiredReplicas: %s\n", string(data))
	data, _ = json.Marshal(h.Config.Spec.Metrics)
	fmt.Printf("Desired Metrics: %s\n", string(data))
	// 遍历每项指标
	for _, oneTargetMtc := range h.Config.Spec.Metrics {
		// 这里oneCurMtc从map拿出来直接就是一个0~1的浮点值了！
		if oneCurMtc, exists := curMetrics.Metrics[oneTargetMtc.Name]; exists {
			// 计算当前指标下的期望的副本数量，计算方式为ceil(当前副本数 * 当前指标值 / 目标指标值)
			expectedReplicas := int(math.Ceil(float64(curReplicaNum) * oneCurMtc / oneTargetMtc.TargetValue))
			// fmt.Printf("in metrics %s, calculate expectRelicas %v\n", oneTargetMtc.Name, expectedReplicas)
			if expectedReplicas > maxDesiredReplicas {
				maxDesiredReplicas = expectedReplicas
			}
		}
	}

	// 确保副本数量在最小和最大值之间
	if maxDesiredReplicas < h.Config.Spec.MinReplicas {
		maxDesiredReplicas = h.Config.Spec.MinReplicas
	} else if maxDesiredReplicas > h.Config.Spec.MaxReplicas {
		maxDesiredReplicas = h.Config.Spec.MaxReplicas
	}

	fmt.Printf("CurReplicaNum: %v, final DesiredReplicas: %v\n", curReplicaNum, maxDesiredReplicas)

	// 返回结果
	return maxDesiredReplicas
}

// 从这个ReplicaSet管理的所有Pods的监控指标百分比中（直接由参数给出），简单加和平均，计算出HPA需要的ReplicaSet级别的监控指标
func CalculateReplicaMetrics(h *HPAType, metrics map[string]PodMetricsEntry) ReplicaMetricsResult {
	result := ReplicaMetricsResult{
		Metrics: make(map[string]float64),
	}

	for _, metric := range h.Config.Spec.Metrics {
		var total float64
		var count float64

		for _, podMetric := range metrics {
			switch metric.Name {
			case "cpu":
				total += podMetric.CPUPercentage
			case "memory":
				total += podMetric.MemoryPercentage
			default:
				fmt.Printf("Unknown metric: %s\n", metric.Name)
				continue
			}
			count++
		}

		if count > 0 {
			result.Metrics[metric.Name] = total / count
		}
	}

	return result
}
