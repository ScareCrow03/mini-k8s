package protocol

// Kind为PingSource的CR，这个对象用于按时间计划触发相应的事件，相应的动作在PingSourceController中实现
type PingSourceSpec struct {
	Schedule string   `yaml:"schedule" json:"schedule"` // 事件触发周期字符串，例如"*/1 * * * *"表示每分钟触发一次
	JsonData string   `yaml:"jsonData" json:"jsonData"` // 事件触发时，写入事件的Json数据
	Sink     SinkType `yaml:"sink" json:"sink"`         // 事件触发时，写入事件的目标；可以是一个Function的标识，也可以是等待被再次解析的JSON串，只要接收者能够对应找到它即可
}

type SinkType struct {
	Ref MetadataType `yaml:"ref" json:"ref"` // 如果希望收到事件的人是一个api对象，那么可以指定它
	URI string       `yaml:"uri" json:"uri"` // 不使用上述方式，可以直接指定一个URI/任何可以找到这个对象的字符串方式，它后续也可以被二次解析
}

// 泛用的事件结构体
type EventType struct {
	EventType string `yaml:"eventType" json:"eventType"` // 事件类型，例如"ping"
	Source    string `yaml:"source" json:"source"`       // 事件来源者，例如"pingSource"的一个容器
	Target    string `yaml:"target" json:"target"`       // 事件目标者，比如可以是Function的标识
	Data      string `yaml:"data" json:"data"`           // 事件数据，应该是字符串，能被后续解析等等
}
