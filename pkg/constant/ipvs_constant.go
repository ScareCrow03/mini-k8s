package constant

var (
	IPTABLES_FILE_PATH = "minik8s_log_dir/iptables_saved.bak"
	IPVS_FILE_PATH     = "minik8s_log_dir/ipvsadm_saved.bak"
	IPSET_FILE_PATH    = "minik8s_log_dir/ipset_saved.bak"

	CLUSTER_CIDR_DEFAULT = "222.111.0.0/16"
)

const (
	// 这是ipvs模式需要的dummy网卡，Service的ClusterIP都会绑定在这个网卡上
	KUBE_DUMMY_INTERFACE_NAME = "mini-kube-ipvs0"

	// 在nat表的PREROUTING和OUTPUT主链中，插入KUBE-SERVICES链
	KUBE_SERVICE_CHAIN_NAME  = "mini-KUBE-SERVICES"
	KUBE_NODEPORT_CHAIN_NAME = "mini-KUBE-NODE-PORT"
	// 用于标记0x10000的链，后续这样的包会在POSTROUTING链中做SNAT
	KUBE_MARK_MASQ_CHAIN_NAME = "mini-KUBE-MARK-MASQ"
	KUBE_MARK_MASQ_VALUE      = "0x10000"

	// 用于标记0x20000的链，后续这样的包会被丢弃
	KUBE_MARK_DROP_CHAIN_NAME = "mini-KUBE-MARK-DROP"
	KUBE_MARK_DROP_VALUE      = "0x20000"

	KUBE_POSTROUTING_CHAIN_NAME = "mini-KUBE-POSTROUTING"

	KUBE_FIREWALL_CHAIN_NAME = "mini-KUBE-FIREWALL"
	KUBE_FORWARD_CHAIN_NAME  = "mini-KUBE-FORWARD"

	// 以下是一些ipset名称
	KUBE_CLUSTER_IP_SET_NAME    = "mini-KUBE-CLUSTER-IP"
	KUBE_NODE_PORT_TCP_SET_NAME = "mini-KUBE-NODE-PORT-TCP"
	KUBE_LOOP_BACK_SET_NAME     = "mini-KUBE-LOOP-BACK"
)
