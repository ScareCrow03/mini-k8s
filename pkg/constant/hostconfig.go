package constant

import "os"

var (
	MasterIp    = os.Getenv("MASTER_IP")
	HttpPreffix = "http://" + MasterIp + ":8080"
	AmqpPath    = "amqp://visitor:123456@" + MasterIp + ":5672/"
	WorkDir     = os.Getenv("WORKDIR")
	NodeName    = os.Getenv("NODENAME")

	// master节点启动一个prometheus，它的配置文件路径；请保证它可读可写，建议chmod 777
	PrometheusConfigPath = "/etc/prometheus/prometheus.yml"
	// 向prometheus发送reload请求的url，默认只在master节点上运行
	PrometheusReloadUrl = "http://" + MasterIp + ":9090/-/reload"

	ServerlessGatewayPrefix = "http://" + MasterIp + ":8050"
)
