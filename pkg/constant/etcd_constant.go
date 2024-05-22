package constant

import "time"

var ( // 生产环境与测试环境的etcd集群IP:port，可以配置；这两个url目前是写死在go源代码中的，后续如果要实际部署到生产环境、可以考虑把它们写入到user_config的yaml文件，在实际启动时再动态读取yaml配置
	EtcdIpPortInProduceEnvDefault = []string{"localhost:2379"}
	EtcdIpPortInTestEnvDefault    = []string{"localhost:2379"}
	// 可以加上其他环境的配置，例如dev等；
	// 连接超时时间
	EtcdTimeout       = 5 * time.Second
	EtcdTestUriPrefix = "/mini-k8s/test" // 测试路径的前缀，这样不需要每次都把整个etcd清空掉；任何改etcd的测试，请认为这个才是你的根目录，之后的路径才是子目录，用法为constant.EtcdTestUriPrefix + "/your_sub_path"
	// 待添加其他可选配置
)

var ( //etcd的key前缀
	EtcdPodPrefix        = "/registry/pod/"
	EtcdKubeletPrefix    = "/registry/kubelet/"
	EtcdServicePrefix    = "/registry/service/"
	EtcdReplicasetPrefix = "/registry/replicaset/"
	EtcdDnsPrefix        = "/registry/dns/"
	EtcdFunctionPrefix   = "/registry/function/"

	EtcdServiceClusterIPPrefix = "/registry/inner/service_cluster-ip/"
)
