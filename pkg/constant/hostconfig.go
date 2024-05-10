package constant

var (
	MasterIp    = "192.168.183.128"
	HttpPreffix = "http://" + MasterIp + ":8080"
	AmqpPath    = "amqp://visitor:123456@" + MasterIp + ":5672/"
)
