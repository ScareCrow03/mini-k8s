package constant

import "time"

var ( // 生产环境与测试环境的etcd集群IP:port，可以配置
	EtcdIpPortInProduceEnv = []string{"localhost:2379"}
	EtcdIpPortInTestEnv    = []string{"localhost:2379"}
	// 可以加上其他环境的配置，例如dev等；
	// 连接超时时间
	EtcdTimeout       = 5 * time.Second
	EtcdTestUriPrefix = "/mini-k8s/test" // 测试路径的前缀，这样不需要每次都把整个etcd清空掉；任何改etcd的测试，请认为这个才是你的根目录，之后的路径才是子目录，用法为constant.EtcdTestUriPrefix + "/your_sub_path"
	// 待添加其他可选配置
)
