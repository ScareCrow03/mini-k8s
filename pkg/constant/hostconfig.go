package constant

import "os"

var (
	MasterIp    = os.Getenv("MASTER_IP")
	HttpPreffix = "http://" + MasterIp + ":8080"
	AmqpPath    = "amqp://visitor:123456@" + MasterIp + ":5672/"
	WorkDir     = os.Getenv("WORKDIR")
	NodeName    = os.Getenv("NODENAME")
)
