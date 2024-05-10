package proxy_server

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/kubeproxy/ipvs_ops"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/protocol/service_cfg"

	"gopkg.in/yaml.v3"
)

// 在这一层可以存储一些必要的状态
type ProxyServer struct {
	IpvsOps *ipvs_ops.IpvsOps
	// 服务名到服务配置的映射
	// 我们这里要求，动态更新的Service信息，必须包含完整的Endpoints信息！每来一个新的Pods，就需要更新所有包含这个Pod的Service的Endpoints信息；然后也能定期做全量替换更新
	// 以下关于Pod更新、同步到SERVICE的eps的方法，在ProxyServer不持有所有的Pods信息的情况下不能正确生效！此时退化为只能使用关于SERVICE自己的前3个方法Add，Delete，UpdateEps，注意此时外面必须传递进来Endpoints状态！
	ServiceMap map[string]*service_cfg.ServiceType
	PodMap     map[string]*protocol.Pod
}

func NewProxyServer(clusterIPCIDR string) *ProxyServer {
	server := &ProxyServer{
		IpvsOps:    ipvs_ops.NewIpvsOps(clusterIPCIDR),
		ServiceMap: make(map[string]*service_cfg.ServiceType),
		PodMap:     make(map[string]*protocol.Pod),
	}
	server.IpvsOps.Init()
	return server
}

// 添加一个新的SERVICE对象
func (ps *ProxyServer) OnServiceAdd(svc *service_cfg.ServiceType) {
	// 这里必须要求Service给出了所有Spec信息
	if len(svc.Status.Endpoints) == 0 {
		// 如果给了一个空的状态，那么由kube-proxy自己到管理的Pods中找，填好eps
		var pods []*protocol.Pod
		for _, pod := range ps.PodMap {
			pods = append(pods, pod)
		}
		fmt.Printf("Selector is %s", svc.Config.Spec.Selector)
		svc.Status.Endpoints = service_cfg.GetEndpointsFromPods(protocol.SelectPodsByLabels(svc.Config.Spec.Selector, pods))
		data, _ := yaml.Marshal(&svc)
		fmt.Printf("Now Service has eps: %s\n", string(data))
	}
	// 添加服务到IPVS
	ps.IpvsOps.AddService(svc)
	// 更新服务映射
	ps.ServiceMap[svc.Config.Metadata.UID] = svc
}

// 删除一个SERVICE对象
func (ps *ProxyServer) OnServiceDelete(svc *service_cfg.ServiceType) {
	// 从IPVS删除服务
	ps.IpvsOps.DelService(svc)
	// 更新服务映射
	delete(ps.ServiceMap, svc.Config.Metadata.UID)
}

// 这个方法仅适用于外面传递进来的service自带endpoints，否则会破坏原有endpoints状态！必须保证这里的old_svc存在，否则请调用OnServiceAdd
func (ps *ProxyServer) OnServiceUpdateEps(old_svc, new_svc *service_cfg.ServiceType) {
	ps.IpvsOps.UpdateServiceEps(old_svc, new_svc)
	ps.ServiceMap[new_svc.Config.Metadata.UID] = new_svc
}

// Pod的添加与删除，影响到相关的Service的Endpoints信息，需要同步一下状态；但是SERVICE的添加和删除，当然不影响其他SERVICE，POD的信息，只需要添加、删除自己的即可
func (ps *ProxyServer) OnPodAdd(pod *protocol.Pod) {
	ps.PodMap[pod.Config.Metadata.UID] = pod
	ps.SyncServicesForAllPods()
}

func (ps *ProxyServer) OnPodDelete(pod *protocol.Pod) {
	delete(ps.PodMap, pod.Config.Metadata.UID)
	ps.SyncServicesForAllPods()
}

// 多Pods信息添加的简单封装，注意这个只作为添加
func (ps *ProxyServer) OnPodsUpdate(pods []*protocol.Pod) {
	for _, pod := range pods {
		ps.PodMap[pod.Config.Metadata.UID] = pod
	}
	ps.SyncServicesForAllPods()
}

// 多Pods信息的同步，这会清空原有的Pods信息！
func (ps *ProxyServer) OnPodsSync(pods []*protocol.Pod) {
	ps.PodMap = make(map[string]*protocol.Pod)
	for _, pod := range pods {
		ps.PodMap[pod.Config.Metadata.UID] = pod
	}
	ps.SyncServicesForAllPods()
}

// 这个函数假定目前kube-proxy掌握所有的Pods信息，那么按这些Pods信息直接同步到所有的Service。如果kube-proxy不掌握所有的Pods信息，不要用这个函数！
func (ps *ProxyServer) SyncServicesForAllPods() {
	updatedSvcs := make(map[string]*service_cfg.ServiceType)

	for _, svc := range ps.ServiceMap {
		if svc.Config.Spec.Selector == nil {
			continue
		}
		// 遍历所有Pod，挑选出能被该SERVICE管理的所有ep
		managed_eps := make([]service_cfg.Endpoint, 0)
		for _, pod := range ps.PodMap {
			if protocol.IsSelectorMatchOnePod(svc.Config.Spec.Selector, pod) {
				new_eps := service_cfg.GetEndpointsFromPod(pod)
				managed_eps = append(managed_eps, new_eps...)
			}
		}
		// 建立一份新的eps保存，注意这里需要深拷贝！
		data, _ := json.Marshal(&svc)
		var svc_copy service_cfg.ServiceType
		json.Unmarshal(data, &svc_copy)

		svc_copy.Status.Endpoints = managed_eps
		updatedSvcs[svc.Config.Metadata.UID] = &svc_copy
	}

	// 然后更新一份新的
	for _, svc := range updatedSvcs {
		ps.IpvsOps.UpdateServiceEps(ps.ServiceMap[svc.Config.Metadata.UID], svc)
		// 写回
		ps.ServiceMap[svc.Config.Metadata.UID] = svc
	}
}
