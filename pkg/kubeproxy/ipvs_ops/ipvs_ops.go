package ipvs_ops

import (
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol/service_cfg"
	"mini-k8s/pkg/utils/net_util"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/coreos/go-iptables/iptables"
	"github.com/moby/ipvs"
	"gopkg.in/yaml.v3"

	"github.com/vishvananda/netlink"
)

// 描述本文件实现的接口
type IpvsOpsInterface interface {
	NewIpvsOps(clusterIPCIDR string)

	// 以下是IpvsOps的成员方法
	Init()
	Clear()
	Close()
	AddService(svc *service_cfg.ServiceType)
	DelService(svc *service_cfg.ServiceType)
	UpdateService(oldSvc *service_cfg.ServiceType, newSvc *service_cfg.ServiceType)

	// 以下两个是方便DEBUG查看的方法，实际应该不会用到
	SaveToFile(iptablesFilePath string, ipvsFilePath string, ipsetFilePath string) error
	RestoreFromFile(iptablesFilePath string, ipvsFilePath string, ipsetFilePath string) error
}

// 这一层只处理给定Service信息后的操作，如果需要保存一些状态，放在kube-proxy的状态结构体中
type IpvsOps struct {
	IptablesClient *iptables.IPTables
	IpvsClient     *ipvs.Handle // 这个handler只在其他地方用到，写入的时候有问题，只能使用Exec
	ClusterIPCIDR  string       // 全空间ClusterIP范围
	// 可以存一些状态量，比如ServiceName到Endpoints的映射
}

func init() {
	// 获取当前用户的主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Failed to get the user home directory:", err)
		return
	}

	constant.IPTABLES_FILE_PATH = filepath.Join(homeDir, constant.IPTABLES_FILE_PATH)
	constant.IPVS_FILE_PATH = filepath.Join(homeDir, constant.IPVS_FILE_PATH)
	constant.IPSET_FILE_PATH = filepath.Join(homeDir, constant.IPSET_FILE_PATH)

	// 创建备份文件的目录
	_ = os.MkdirAll(filepath.Dir(constant.IPTABLES_FILE_PATH), 0755)
	// 文件本身不需要创建，因为每次都是由命令行输出重定向到文件
}

func NewIpvsOps(clusterIPCIDR string) *IpvsOps {
	ops := &IpvsOps{
		ClusterIPCIDR: clusterIPCIDR,
	}
	ops.IptablesClient, _ = iptables.New()
	ops.IpvsClient, _ = ipvs.New("")

	return ops
}

func (ops *IpvsOps) Close() {
	ops.IpvsClient.Close()
	ops.IptablesClient = nil
	ops.IpvsClient = nil
}

func (ops *IpvsOps) Init() {
	// 创建ipvs模式需要的dummy网卡
	createDummyInterface(constant.KUBE_DUMMY_INTERFACE_NAME)

	// 创建表与链名的添加关系
	iptTable2Chains := map[string][]string{}
	iptTable2Chains["nat"] = []string{
		constant.KUBE_SERVICE_CHAIN_NAME,
		constant.KUBE_NODEPORT_CHAIN_NAME,
		constant.KUBE_MARK_MASQ_CHAIN_NAME,
		constant.KUBE_MARK_DROP_CHAIN_NAME,
		constant.KUBE_POSTROUTING_CHAIN_NAME,
		constant.KUBE_FIREWALL_CHAIN_NAME,
	}

	iptTable2Chains["filter"] = []string{
		constant.KUBE_FIREWALL_CHAIN_NAME,
		constant.KUBE_FORWARD_CHAIN_NAME,
	}

	// 创建链，不需要检查是否存在，多次创建幂等
	for table, chains := range iptTable2Chains {
		for _, chain := range chains {
			_ = ops.IptablesClient.NewChain(table, chain)
		}
	}

	// 创建ipset集合，注意每个类型都不同
	// 第一个KUBE-CLUSTER-IP是ClusterIP:port的集合
	// 第二个KUBE-NODE-PORT-TCP是NodePort tcp的集合，为了简单我们只管tcp
	// 第三个KUBE-LOOP-BACK存放endpoints信息，包含了每个Service内PodIP:PodPort:PodIP三元组；为什么是这样？因为这个ipset在起作用时已经是在POSTROUTING链中了，它已经由ipvs做好了DNAT，此时的dstIP:dstPort就是目标的PodIP:PodPort！这个ipset被建立起来是为了解决某个Pod访问自己所属的Service后，后续流量又到了自己的情况
	// 直接创建出来，不做检查，应该保证命令行输入正确即可
	cmd := exec.Command("ipset", "create", constant.KUBE_CLUSTER_IP_SET_NAME, "hash:ip,port")
	_ = cmd.Run()

	cmd = exec.Command("ipset", "create", constant.KUBE_NODE_PORT_TCP_SET_NAME, "bitmap:port", "range", "0-65535")
	output, err := cmd.Output()
	if err != nil {
		logger.KError("Failed to execute command: %v, output: %s", err, output)
	}

	cmd = exec.Command("ipset", "create", constant.KUBE_LOOP_BACK_SET_NAME, "hash:ip,port,ip")
	_ = cmd.Run()

	// 初始化各iptables规则
	// 对于PREROUTING和OUTPUT主链，各自添加无条件跳转到KUBE-SERVICES链的规则
	ops.IptablesClient.AppendUnique("nat", "PREROUTING", "-j", constant.KUBE_SERVICE_CHAIN_NAME, "-m", "comment", "--comment", "mini-k8s service portals")
	ops.IptablesClient.AppendUnique("nat", "OUTPUT", "-j", constant.KUBE_SERVICE_CHAIN_NAME, "-m", "comment", "--comment", "mini-k8s service portals")

	// 添加nat表的KUBE-SERVICE链规则
	// 第一，如果srcIP不在ClusterIP范围内，而且dstIP:dstPort符合ClusterIP:port的访问，那么跳转到KUBE-MARK-MASQ，打上标记
	ops.IptablesClient.AppendUnique("nat", constant.KUBE_SERVICE_CHAIN_NAME, "!", "-s", ops.ClusterIPCIDR, "-m", "set", "--match-set", constant.KUBE_CLUSTER_IP_SET_NAME, "dst,dst", "-j", constant.KUBE_MARK_MASQ_CHAIN_NAME, "-m", "comment", "--comment", "mini-k8s service cluster ip + port for masquerade purpose")

	// 第二，dstIP是对本地的访问，那么跳转到KUBE-NODE-PORT链
	ops.IptablesClient.AppendUnique("nat", constant.KUBE_SERVICE_CHAIN_NAME, "-m", "addrtype", "--dst-type", "LOCAL", "-j", constant.KUBE_NODEPORT_CHAIN_NAME)

	// 第三，如果符合dstIP:dstPort符合ClusterIP:port的访问，则接受它
	ops.IptablesClient.AppendUnique("nat", constant.KUBE_SERVICE_CHAIN_NAME, "-m", "set", "--match-set", constant.KUBE_CLUSTER_IP_SET_NAME, "dst,dst", "-j", "ACCEPT")

	// 添加KUBE-MARK-MASQ链的规则，只需要打上0x10000标记
	ops.IptablesClient.AppendUnique("nat", constant.KUBE_MARK_MASQ_CHAIN_NAME, "-j", "MARK", "--set-xmark", constant.KUBE_MARK_MASQ_VALUE+"/"+constant.KUBE_MARK_MASQ_VALUE)

	// 添加KUBE-NODE-PORT链的规则
	// 如果符合dstPort是NodePort的访问，那么跳转到KUBE_MARK_MASQ打上0x10000标记
	ops.IptablesClient.AppendUnique("nat", constant.KUBE_NODEPORT_CHAIN_NAME, "-m", "set", "--match-set", constant.KUBE_NODE_PORT_TCP_SET_NAME, "dst", "-j", constant.KUBE_MARK_MASQ_CHAIN_NAME, "-m", "comment", "--comment", "mini-k8s nodeport TCP port for masquerade purpose")

	// 添加POSTROUTING主链的规则，无条件跳转到KUBE-POSTROUTING链
	ops.IptablesClient.AppendUnique("nat", "POSTROUTING", "-j", constant.KUBE_POSTROUTING_CHAIN_NAME, "-m", "comment", "--comment", "mini-k8s postrouting rules")

	// 添加KUBE-POSTROUTING链的规则，此处已经由ipvs做好了DNAT，这里需要做的是对于有0x10000标记的包，做SNAT；如果这个包没有被打上0x10000标记，不管它，直接返回即可
	// 第一，如果匹配上KUBE-LOOP-BACK这个ipset中的dstIP,dstPort,srcIP三元组，那么采取MASQUERADE这个动作，按照当前网卡的ip做SNAT
	ops.IptablesClient.AppendUnique("nat", constant.KUBE_POSTROUTING_CHAIN_NAME, "-m", "set", "--match-set", constant.KUBE_LOOP_BACK_SET_NAME, "dst,dst,src", "-j", "MASQUERADE", "-m", "comment", "--comment", "mini-k8s endpoints dst ip:port, source ip for solving hairpin purpose")

	// 第二，如果不在上述ipset内，而且没有被打上0x10000标记，那么直接返回
	ops.IptablesClient.AppendUnique("nat", constant.KUBE_POSTROUTING_CHAIN_NAME, "-m", "mark", "!", "--mark", constant.KUBE_MARK_MASQ_VALUE+"/"+constant.KUBE_MARK_MASQ_VALUE, "-j", "RETURN")

	// 第三，如果不在上述ipset内，而且被打上0x10000标记，那么先通过xor把标记去掉（因为可能发到其他Node上了），做紧接着第四条规则无条件做MASQUERADE，也即SNAT
	ops.IptablesClient.AppendUnique("nat", constant.KUBE_POSTROUTING_CHAIN_NAME, "-j", "MARK", "--xor-mark", constant.KUBE_MARK_MASQ_VALUE)
	ops.IptablesClient.AppendUnique("nat", constant.KUBE_POSTROUTING_CHAIN_NAME, "-j", "MASQUERADE", "-m", "comment", "--comment", "mini-k8s service traffic requiring SNAT")

	// 原始k8s中还添加KUBE-FIREWALL链的规则，目前可能是无关紧要的，故暂时不写
}

func (ops *IpvsOps) Clear() { // 只删除必要的部分！
	ops.IptablesClient.ClearChain("nat", constant.KUBE_SERVICE_CHAIN_NAME)
	ops.IptablesClient.ClearChain("nat", constant.KUBE_NODEPORT_CHAIN_NAME)
	ops.IptablesClient.ClearChain("nat", constant.KUBE_MARK_MASQ_CHAIN_NAME)
	ops.IptablesClient.ClearChain("nat", constant.KUBE_MARK_DROP_CHAIN_NAME)
	ops.IptablesClient.ClearChain("nat", constant.KUBE_POSTROUTING_CHAIN_NAME)
	ops.IptablesClient.ClearChain("nat", constant.KUBE_FIREWALL_CHAIN_NAME)
	ops.IptablesClient.ClearChain("nat", constant.KUBE_FORWARD_CHAIN_NAME)

	ops.IptablesClient.ClearChain("filter", constant.KUBE_FIREWALL_CHAIN_NAME)
	ops.IptablesClient.ClearChain("filter", constant.KUBE_FORWARD_CHAIN_NAME)

	// 清除所有ipset
	ipsetSets := []string{
		"mini-KUBE-CLUSTER-IP",
		"mini-KUBE-NODE-PORT-TCP",
		"mini-KUBE-LOOP-BACK",
	}

	for _, set := range ipsetSets {
		cmd := exec.Command("ipset", "flush", set)
		_ = cmd.Run()
		// 删除set
		cmd = exec.Command("ipset", "destroy", set)
		_ = cmd.Run()
	}

	// 清除所有ipvs规则
	ops.IpvsClient.Flush()

	// 清除dummy网卡绑定的所有IP
	clearAllIPsFromDummyInterface(constant.KUBE_DUMMY_INTERFACE_NAME)
}

// 添加一个新的Service、配置相关的iptables, ipvs, ipset（考虑这与Update有什么区别？如果没有区别应该可以直接复用）
func (ops *IpvsOps) AddService(svc *service_cfg.ServiceType) {
	// 将clusterIP绑定到dummy网卡
	bindClusterIPToDummyInterface(constant.KUBE_DUMMY_INTERFACE_NAME, svc.Config.Spec.ClusterIP)

	// 添加ClusterIP:port到KUBE-CLUSTER-IP这个ipset
	for _, port := range svc.Config.Spec.Ports {
		cmd := exec.Command("ipset", "add", constant.KUBE_CLUSTER_IP_SET_NAME, svc.Config.Spec.ClusterIP+",tcp:"+fmt.Sprint(port.Port))
		fmt.Printf(cmd.String() + "\n")
		err := cmd.Run()
		if err != nil {
			logger.KError("Failed to add clusterIP %s:%d to ipset %s: %v", svc.Config.Spec.ClusterIP, port.Port, constant.KUBE_CLUSTER_IP_SET_NAME, err)
		}
	}

	// 如果需要，添加NodePort到KUBE-NODE-PORT-TCP这个ipset
	if svc.Config.Spec.Type == service_cfg.SERVICE_TYPE_NODEPORT_STR {
		for _, port := range svc.Config.Spec.Ports {
			cmd := exec.Command("ipset", "add", constant.KUBE_NODE_PORT_TCP_SET_NAME, fmt.Sprint(port.NodePort))
			err := cmd.Run()
			if err != nil {
				logger.KError("Failed to add nodePort %d to ipset %s: %v", port.NodePort, constant.KUBE_NODE_PORT_TCP_SET_NAME, err)
			}
		}
	}

	// 添加Endpoints到KUBE-LOOP-BACK这个ipset
	for _, ep := range svc.Status.Endpoints {
		cmd := exec.Command("ipset", "add", constant.KUBE_LOOP_BACK_SET_NAME, ep.IP+",tcp:"+fmt.Sprint(ep.Port)+","+ep.IP)
		fmt.Printf(cmd.String() + "\n")
		err := cmd.Run()
		if err != nil {
			logger.KError("Failed to add endpoint %s:%s:%s to ipset %s: %v", ep.IP, "tcp:"+fmt.Sprint(ep.Port), ep.IP, constant.KUBE_LOOP_BACK_SET_NAME, err)
		}
	}

	// 以下这两个函数或许可以提取出来复用？

	// 添加ClusterIP的ipvs规则，是关于ClusterIP:port的DNAT规则，对应到符合相应targetPort暴露的PodIP:PodPort
	clusterIP := svc.Config.Spec.ClusterIP
	ports := svc.Config.Spec.Ports

	for _, port := range ports {
		// 创建一个新的IPVS服务，具有ClusterIP:port，这是前端对应的虚服务

		// // 添加IPVS服务
		svc_clusterip_addr := clusterIP + fmt.Sprintf(":%v", port.Port)
		_, err := exec.Command("ipvsadm", "-A", "-t", svc_clusterip_addr, "-s", "rr").Output()

		if err != nil {
			logger.KError("Failed to add IPVS service for %s:%d, reason: %v", clusterIP, port.Port, err)
			continue
		}

		// 为每一个匹配选择器的Pod创建一个IPVS规则
		for _, endpoint := range svc.Status.Endpoints {
			// 需要保证PodPort与TargetPort一致，才对应于这个前端虚服务添加IPVS目标
			if port.TargetPort != endpoint.Port {
				continue
			}
			// // 创建一个新的IPVS目标，具有PodIP:PodPort
			// // 添加IPVS目标
			ep_str := endpoint.IP + fmt.Sprintf(":%v", endpoint.Port)
			_, err := exec.Command("ipvsadm", "-a", "-t", svc_clusterip_addr, "-r", ep_str, "-m").Output()
			if err != nil {
				logger.KError("Failed to add IPVS destination for %s:%d: %v", endpoint.IP, endpoint.Port, err)
			}
		}
	}

	// 添加NodePort的ipvs规则，是关于NodePort的DNAT规则，对应到符合相应targetPort暴露的PodIP:PodPort
	// 这里要求必须有NodePort指定
	if svc.Config.Spec.Type == service_cfg.SERVICE_TYPE_NODEPORT_STR {
		nodeIP, _ := net_util.GetNodeIP()
		for _, port := range ports { // 也许有一些服务没有暴露NodePort
			if port.NodePort == 0 {
				continue
			}
			// 根据一项Service层面暴露Port的规则，创建一个新的IPVS服务，具有NodeIP:NodePort，这是前端对应的虚服务
			// // 添加IPVS服务
			svc_nodeport_addr := nodeIP + fmt.Sprintf(":%v", port.NodePort)
			_, err := exec.Command("ipvsadm", "-A", "-t", svc_nodeport_addr, "-s", "rr").Output()
			if err != nil {
				logger.KError("Failed to add IPVS service for %s:%d: %v", nodeIP, port.NodePort, err)
				continue
			}

			// 为每一个匹配选择器的Pod创建一个IPVS规则
			for _, endpoint := range svc.Status.Endpoints {
				// 需要保证PodPort与TargetPort一致，才对应于这个前端虚服务添加IPVS目标
				if port.TargetPort != endpoint.Port {
					continue
				}
				// // 创建一个新的IPVS目标，具有PodIP:PodPort

				// // 添加IPVS目标
				// err := ops.IpvsClient.NewDestination(ipvsSvc, ipvsDst)
				ep_str := endpoint.IP + fmt.Sprintf(":%v", endpoint.Port)
				_, err := exec.Command("ipvsadm", "-a", "-t", svc_nodeport_addr, "-r", ep_str, "-m").Output()
				if err != nil {
					logger.KError("Failed to add IPVS destination for %s:%d: %v", endpoint.IP, endpoint.Port, err)
				}
			}
		}
	}

}

// 添加的逆过程
func (ops *IpvsOps) DelService(svc *service_cfg.ServiceType) {
	// 解绑ClusterIP
	unbindClusterIPFromDummyInterface(constant.KUBE_DUMMY_INTERFACE_NAME, svc.Config.Spec.ClusterIP)

	// 从KUBE-CLUSTER-IP这个ipset中删除ClusterIP:port
	for _, port := range svc.Config.Spec.Ports {
		cmd := exec.Command("ipset", "del", constant.KUBE_CLUSTER_IP_SET_NAME, svc.Config.Spec.ClusterIP+",tcp:"+fmt.Sprint(port.Port))
		fmt.Printf(cmd.String() + "\n")
		err := cmd.Run()
		if err != nil {
			logger.KError("Failed to delete clusterIP %s:%d from ipset %s: %v", svc.Config.Spec.ClusterIP, port.Port, constant.KUBE_CLUSTER_IP_SET_NAME, err)
		}
	}

	// 如果需要，从KUBE-NODE-PORT-TCP这个ipset中删除NodePort
	if svc.Config.Spec.Type == service_cfg.SERVICE_TYPE_NODEPORT_STR {
		for _, port := range svc.Config.Spec.Ports {
			cmd := exec.Command("ipset", "del", constant.KUBE_NODE_PORT_TCP_SET_NAME, fmt.Sprint(port.NodePort))
			err := cmd.Run()
			if err != nil {
				logger.KError("Failed to delete nodePort %d from ipset %s: %v", port.NodePort, constant.KUBE_NODE_PORT_TCP_SET_NAME, err)
			}
		}
	}

	// 从KUBE-LOOP-BACK这个ipset中删除Endpoints
	for _, ep := range svc.Status.Endpoints {
		cmd := exec.Command("ipset", "del", constant.KUBE_LOOP_BACK_SET_NAME, ep.IP+",tcp:"+fmt.Sprint(ep.Port)+","+ep.IP)
		fmt.Printf(cmd.String() + "\n")
		err := cmd.Run()
		if err != nil {
			logger.KError("Failed to delete endpoint %s:%s:%s from ipset %s: %v", ep.IP, "tcp:"+fmt.Sprint(ep.Port), ep.IP, constant.KUBE_LOOP_BACK_SET_NAME, err)
		}
	}

	// 删除关于ClusterIP:port的DNAT规则，此处删除只需要指定ClusterIP:port，而无需对应的Endpoints
	clusterIP := svc.Config.Spec.ClusterIP
	ports := svc.Config.Spec.Ports

	for _, port := range ports {
		svc_clusterip_addr := clusterIP + fmt.Sprintf(":%v", port.Port)
		_, err := exec.Command("ipvsadm", "-D", "-t", svc_clusterip_addr).Output()

		if err != nil {
			logger.KError("Failed to delete IPVS service for %s:%d, reason: %v", clusterIP, port.Port, err)
			continue
		}
	}

	// 删除关于NodePort的DNAT规则，对应到符合相应targetPort暴露的PodIP:PodPort
	if svc.Config.Spec.Type == service_cfg.SERVICE_TYPE_NODEPORT_STR {
		nodeIP, _ := net_util.GetNodeIP()
		for _, port := range ports {
			if port.NodePort == 0 {
				continue
			}

			svc_nodeport_addr := nodeIP + fmt.Sprintf(":%v", port.NodePort)
			_, err := exec.Command("ipvsadm", "-D", "-t", svc_nodeport_addr).Output()
			if err != nil {
				logger.KError("Failed to delete IPVS service for %s:%d: %v", nodeIP, port.NodePort, err)
				continue
			}
		}
	}
}

// 更新Service配置，要求必须是同一个Service对象，只是endpoints发生了改变！这是很细粒度的操作，不会涉及到ClusterIP、NodePort的变化
func (ops *IpvsOps) UpdateServiceEps(oldSvc, newSvc *service_cfg.ServiceType) {
	addedEndpoints, removedEndpoints := service_cfg.CompareEndpoints(oldSvc.Status.Endpoints, newSvc.Status.Endpoints)

	// 打印状态
	data, _ := yaml.Marshal(&oldSvc)
	fmt.Printf("oldSvc status: %s", string(data))

	data, _ = yaml.Marshal(&newSvc)
	fmt.Printf("newSvc status: %s", string(data))

	// 反向映射targetPort到ServicePort
	clusterIP := newSvc.Config.Spec.ClusterIP
	nodeIP, _ := net_util.GetNodeIP()

	targetPortMap := make(map[int]service_cfg.ServicePort)
	for _, port := range newSvc.Config.Spec.Ports {
		targetPortMap[port.TargetPort] = port
	}

	// endpoints只需要通过targetPort的对应反向索引到Cluster消息即可！
	for _, ep := range removedEndpoints {
		// 从KUBE-LOOP-BACK这个ipset中删除Endpoint
		cmd := exec.Command("ipset", "del", constant.KUBE_LOOP_BACK_SET_NAME, ep.IP+",tcp:"+fmt.Sprint(ep.Port)+","+ep.IP)
		err := cmd.Run()
		if err != nil {
			logger.KError("Failed to delete endpoint %s:%s:%s from ipset %s: %v", ep.IP, "tcp:"+fmt.Sprint(ep.Port), ep.IP, constant.KUBE_LOOP_BACK_SET_NAME, err)
		}

		// 在ipvs删除ClusterIP:port关于这个ep的DNAT规则
		if targetPortMap[ep.Port] != (service_cfg.ServicePort{}) {
			svc_clusterip_addr := clusterIP + fmt.Sprintf(":%v", targetPortMap[ep.Port].Port)
			ep_str := ep.IP + fmt.Sprintf(":%v", ep.Port)
			_, _ = exec.Command("ipvsadm", "-d", "-t", svc_clusterip_addr, "-r", ep_str).Output()

			// 如果它具有NodePort，一并删掉
			if targetPortMap[ep.Port].NodePort != 0 {
				svc_nodeport_addr := nodeIP + fmt.Sprintf(":%v", targetPortMap[ep.Port].NodePort)
				_, _ = exec.Command("ipvsadm", "-d", "-t", svc_nodeport_addr, "-r", ep_str).Output()
			}
		}
	}

	// 添加新的endpoints
	for _, ep := range addedEndpoints {
		// 添加到KUBE-LOOP-BACK这个ipset中
		cmd := exec.Command("ipset", "add", constant.KUBE_LOOP_BACK_SET_NAME, ep.IP+",tcp:"+fmt.Sprint(ep.Port)+","+ep.IP)
		_ = cmd.Run()

		// 在ipvs添加ClusterIP:port关于这个ep的DNAT规则
		if targetPortMap[ep.Port] != (service_cfg.ServicePort{}) {
			svc_clusterip_addr := clusterIP + fmt.Sprintf(":%v", targetPortMap[ep.Port].Port)
			ep_str := ep.IP + fmt.Sprintf(":%v", ep.Port)
			_, _ = exec.Command("ipvsadm", "-a", "-t", svc_clusterip_addr, "-r", ep_str, "-m").Output()

			// 如果它具有NodePort，一并添加
			if targetPortMap[ep.Port].NodePort != 0 {
				svc_nodeport_addr := nodeIP + fmt.Sprintf(":%v", targetPortMap[ep.Port].NodePort)
				_, _ = exec.Command("ipvsadm", "-a", "-t", svc_nodeport_addr, "-r", ep_str, "-m").Output()
			}
		}
	}

}

// 执行以下命令行时需要sudo权限
func (ops *IpvsOps) SaveToFile(iptablesFilePath string, ipvsFilePath string, ipsetFilePath string) error {
	if iptablesFilePath != "" {
		// 保存iptables配置
		iptablesCmd := exec.Command("sh", "-c", "iptables-save > "+iptablesFilePath)
		err := iptablesCmd.Run()
		if err != nil {
			logger.KError("Failed to save iptables config: %v", err)
			return err
		}
	}

	if ipvsFilePath != "" {
		// 保存ipvs配置
		ipvsCmd := exec.Command("sh", "-c", "ipvsadm -S > "+ipvsFilePath)
		err := ipvsCmd.Run()
		if err != nil {
			logger.KError("Failed to save ipvs config: %v", err)
			return err
		}
	}

	if ipsetFilePath != "" {
		// 保存ipset配置
		ipsetCmd := exec.Command("sh", "-c", "ipset save > "+ipsetFilePath)
		err := ipsetCmd.Run()
		if err != nil {
			logger.KError("Failed to save ipset config: %v", err)
			return err
		}
	}

	return nil
}

func (ops *IpvsOps) RestoreFromFile(iptablesFilePath string, ipvsFilePath string, ipsetFilePath string) error {
	// 恢复iptables配置
	if ipsetFilePath != "" {
		iptablesCmd := exec.Command("sh", "-c", "iptables-restore < "+iptablesFilePath)
		err := iptablesCmd.Run()
		if err != nil {
			logger.KError("Failed to restore iptables config: %v", err)
			return err
		}
	}

	// 恢复ipvs配置
	if ipvsFilePath != "" {
		ipvsCmd := exec.Command("sh", "-c", "ipvsadm -R < "+ipvsFilePath)
		err := ipvsCmd.Run()
		if err != nil {
			logger.KError("Failed to restore ipvs config: %v", err)
			return err
		}
	}

	// 恢复ipset配置
	if ipsetFilePath != "" {
		ipsetCmd := exec.Command("sh", "-c", "ipset restore < "+ipsetFilePath)
		err := ipsetCmd.Run()
		if err != nil {
			logger.KError("Failed to restore ipset config: %v", err)
			return err
		}
	}
	return nil
}

// 创建ipvs模式需要的dummy网卡设备
func createDummyInterface(name string) error {
	if name == "" {
		return nil
	}

	if _, err := netlink.LinkByName(name); err == nil {
		// 如果dummy网卡已经存在，那么不需要再创建
		logger.KInfo("Dummy interface %s already exists, no need to create", name)
		return nil
	}

	// 创建一个新的dummy网卡
	dummy := &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{
			Name: name,
		},
	}

	// 添加dummy网卡到系统
	err := netlink.LinkAdd(dummy)
	if err != nil {
		logger.KError("Failed to add dummy interface %s: %v", name, err)
		return err
	}

	// 启动dummy网卡
	err = netlink.LinkSetUp(dummy)
	if err != nil {
		logger.KError("Failed to set up dummy interface %s: %v", name, err)
		return err
	}

	return nil
}

// 绑定ClusterIP到dummy网卡
func bindClusterIPToDummyInterface(name string, clusterIP string) error {
	// 获取dummy网卡
	link, err := netlink.LinkByName(name)
	if err != nil {
		logger.KError("Failed to get dummy interface %s: %v", name, err)
		return err
	}

	// 获取dummy网卡的所有IP地址
	addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		logger.KError("Failed to get IP addresses of dummy interface %s: %v", name, err)
		return err
	}

	// 检查ClusterIP是否已经绑定到dummy网卡
	for _, addr := range addrs {
		if addr.IPNet.String() == clusterIP { // 如果已绑定，OK
			return nil
		}
	}

	// 绑定ClusterIP到dummy网卡
	addr, err := netlink.ParseAddr(clusterIP + "/32")
	if err != nil {
		logger.KError("Failed to parse ClusterIP %s: %v", clusterIP, err)
		return err
	}
	err = netlink.AddrAdd(link, addr)
	if err != nil {
		logger.KError("Failed to bind ClusterIP %s to dummy interface %s: %v", clusterIP, name, err)
		return err
	}

	return nil
}

// 解绑
func unbindClusterIPFromDummyInterface(name string, clusterIP string) error {
	// 获取dummy网卡
	link, err := netlink.LinkByName(name)
	if err != nil {
		logger.KError("Failed to get dummy interface %s: %v", name, err)
	}
	// 获取地址
	addr, err := netlink.ParseAddr(clusterIP + "/32")
	if err != nil {
		logger.KError("Failed to parse ClusterIP %s: %v", clusterIP, err)
	}

	// 删除
	err = netlink.AddrDel(link, addr)
	if err != nil {
		logger.KError("Failed to unbind ClusterIP %s from dummy interface %s: %v", clusterIP, name, err)
		return err
	}
	return nil
}

// 清理dummy网卡上的所有IP
func clearAllIPsFromDummyInterface(name string) error {
	// 获取dummy网卡
	link, err := netlink.LinkByName(name)
	if err != nil {
		logger.KError("Failed to get dummy interface %s: %v", name, err)
		return err
	}

	// 获取dummy网卡的所有IP地址
	addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		logger.KError("Failed to get IP addresses of dummy interface %s: %v", name, err)
		return err
	}

	// 遍历所有IP地址并移除
	for _, addr := range addrs {
		err = netlink.AddrDel(link, &addr)
		if err != nil {
			logger.KError("Failed to remove IP %s from dummy interface %s: %v", addr.IPNet.String(), name, err)
			return err
		}
	}

	return nil
}
