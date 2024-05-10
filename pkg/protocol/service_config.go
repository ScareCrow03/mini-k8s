package protocol

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ServicePort struct { // 一个Port映射关系
	Name       string `yaml:"name" json:"name"`
	Port       int    `yaml:"port" json:"port"`             // 对外暴露的端口
	TargetPort int    `yaml:"targetPort" json:"targetPort"` // 对pod暴露的端口
	NodePort   int    `yaml:"nodePort" json:"nodePort"`     // 仅当Type为NodePort时适用；允许为空，此时由api-server随机分配30000-32767之间一个未被占用的（到达kube-proxy时，这个消息必须已经确定好了！），手动指定的应该也需要在上述范围内；只有一组Port-TargetPort对应好之后，NodePort才有意义，即使用<NodeIP>:<NodePort>访问Service的效果，与本服务访问<ClusterIP>:<Port>的效果一致
	Protocol   string `yaml:"protocol" json:"protocol"`     // 默认是TCP
}

type ServiceSpecType struct {
	Type     string            `yaml:"type" json:"type"` // Service的类型，可以是ClusterIP或NodePort；如果是NodePort
	Ports    []ServicePort     `yaml:"ports" json:"ports"`
	Selector map[string]string `yaml:"selector" json:"selector"`

	ClusterIP string `yaml:"clusterIP" json:"clusterIP"` // Service在集群内部的静态CLUSTER_IP地址
}

type Endpoint struct {
	PodUID string `yaml:"podUID" json:"podUID"`
	IP     string `yaml:"ip" json:"ip"`
	Port   int    `yaml:"port" json:"port"`
}

type ServiceStatus struct {
	Endpoints []Endpoint `yaml:"endpoints" json:"endpoints"` // Endpoints字段记录了满足Service选择器条件的Pod的IP地址和端口；注意是并集！
}

type ServiceConfig struct {
	ApiVersion string          `yaml:"apiVersion" json:"apiVersion"`
	Kind       string          `yaml:"kind" json:"kind"`
	Metadata   MetadataType    `yaml:"metadata" json:"metadata"`
	Spec       ServiceSpecType `yaml:"spec" json:"spec"`
}

const (
	SERVICE_TYPE_CLUSTERIP_STR = "ClusterIP"
	SERVICE_TYPE_NODEPORT_STR  = "NodePort"
)

type ServiceType struct {
	Config ServiceConfig `yaml:"config" json:"config"`
	Status ServiceStatus `yaml:"status" json:"status"`
}

func GetEndpointsFromPods(pods []*Pod) []Endpoint {
	eps := make([]Endpoint, 0)
	for _, pod := range pods {
		data, _ := yaml.Marshal(&pod)
		fmt.Printf("Now pod is: %s\n", string(data))
		for _, container := range pod.Config.Spec.Containers {
			// 如果这个容器暴露的端口不为空
			if len(container.Ports) > 0 {
				// 逐一绑定
				for _, port := range container.Ports {
					ep := Endpoint{
						PodUID: pod.Config.Metadata.UID,
						IP:     pod.Status.IP,
						Port:   int(port.ContainerPort),
					}
					data, _ := yaml.Marshal(&ep)
					fmt.Printf("Now ep is: %s\n", string(data))
					eps = append(eps, ep)
				}
			}
		}
	}
	return eps
}

func GetEndpointsFromPod(pod *Pod) []Endpoint {
	eps := make([]Endpoint, 0)
	for _, container := range pod.Config.Spec.Containers {
		// 如果这个容器暴露的端口不为空
		if len(container.Ports) > 0 {
			// 逐一绑定
			for _, port := range container.Ports {
				ep := Endpoint{
					PodUID: pod.Config.Metadata.UID,
					IP:     pod.Status.IP,
					Port:   int(port.ContainerPort),
				}
				eps = append(eps, ep)
			}
		}
	}
	return eps
}

func CompareEndpoints(oldEndpoints, newEndpoints []Endpoint) (added, removed []Endpoint) {
	oldMap := make(map[string]bool)
	newMap := make(map[string]bool)

	for _, ep := range oldEndpoints {
		oldMap[ep.IP+":"+fmt.Sprint(ep.Port)] = true
	}

	for _, ep := range newEndpoints {
		newMap[ep.IP+":"+fmt.Sprint(ep.Port)] = true
		if !oldMap[ep.IP+":"+fmt.Sprint(ep.Port)] {
			added = append(added, ep)
		}
	}

	for _, ep := range oldEndpoints {
		if !newMap[ep.IP+":"+fmt.Sprint(ep.Port)] {
			removed = append(removed, ep)
		}
	}

	return added, removed
}
