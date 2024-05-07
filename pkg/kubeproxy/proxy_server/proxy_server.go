package proxy_server

import (
	"mini-k8s/pkg/kubeproxy/ipvs_ops"
	"mini-k8s/pkg/protocol/service_cfg"
)

// 在这一层可以存储一些必要的状态
type ProxyServer struct {
	IpvsOps *ipvs_ops.IpvsOps
	// 服务名到服务配置的映射
	ServiceMap map[string]*service_cfg.ServiceType
}

func NewProxyServer(clusterIPCIDR string) *ProxyServer {
	return &ProxyServer{
		IpvsOps:    ipvs_ops.NewIpvsOps(clusterIPCIDR),
		ServiceMap: make(map[string]*service_cfg.ServiceType),
	}
}

// 要求传递进来的add信息中已经具有完整的Service信息，包含Endpoints状态！
// Endpoints认为是与Service紧密联系的，不应该单独处理；api-server需要感知到某个Pod发生变动时，反向索引到所有包含它的Service，然后更新这些Service的Endpoints状态，然后通知kube-proxy作出更新
// 注意这里的Endpoints状态含有PodIP，是动态信息，而没有yaml中静态指定；可能还需要绕一下，例如kubelet定期向api-server汇报Pod状态，此时才具有PodIP信息
func (ps *ProxyServer) OnServiceAdd(svc *service_cfg.ServiceType) {
	// 添加服务到IPVS
	ps.IpvsOps.AddService(svc)
	// 更新服务映射
	ps.ServiceMap[svc.Config.Metadata.Name] = svc
}

func (ps *ProxyServer) OnServiceDelete(svc *service_cfg.ServiceType) {
	// 从IPVS删除服务
	ps.IpvsOps.DelService(svc)
	// 更新服务映射
	delete(ps.ServiceMap, svc.Config.Metadata.Name)
}

func (ps *ProxyServer) OnServiceUpdate(oldSvc, newSvc *service_cfg.ServiceType) {
	// 更新IPVS服务
	ps.IpvsOps.UpdateService(oldSvc, newSvc)
	// 更新服务映射
	ps.ServiceMap[newSvc.Config.Metadata.Name] = newSvc
}
