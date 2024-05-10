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
	ServiceName string `yaml:"svcName" json:"svcName"`
	ServiceIp   string `yaml:"svcIp" json:"svcIp"`
	Port        int    `yaml:"port" json:"port"`
}
