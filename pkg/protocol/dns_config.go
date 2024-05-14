package protocol

type Dns struct {
	Basic `yaml:",inline" json:",inline"`
	Spec  DnsSpec `yaml:"spec" json:"spec"`
}

type DnsSpec struct {
	Host  string `yaml:"host" json:"host"`
	Paths []Path `yaml:"paths" json:"paths"`
}

type Path struct {
	SubPath     string `yaml:"subPath" json:"subPath"`
	ServiceName string `yaml:"serviceName" json:"serviceName"`
	ServiceIp   string `yaml:"serviceIp" json:"serviceIp"`
	Port        int    `yaml:"port" json:"port"`
}

type DnsMsg struct {
	Dns        Dns
	HostConfig []string
}
