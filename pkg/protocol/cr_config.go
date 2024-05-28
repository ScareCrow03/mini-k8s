package protocol

import "strings"

// 这个CRType用于处理用户自定义资源，yaml文件中要有前三个顶级字段，然后Spec顶级字段之下可以是任何东西
// 为了把泛型Spec解析到具体类型，请使用pkg/utils/type_cast包中的GetObjectFromInterface方法
type CRType struct {
	ApiVersion string       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Metadata   MetadataType `yaml:"metadata" json:"metadata"`
	Spec       interface{}  `yaml:"spec" json:"spec"`
}

// 允许用户传递简写的Kind字段，然后在这里进行扩展成完全形式，这个一般不会被用到，除非是kubectl想偷懒
func DoCRKindStrExpands(kindStr *string) {
	*kindStr = strings.ToLower(*kindStr)
	if strings.Contains(*kindStr, "ping") {
		*kindStr = "pingsource"
	}

	if strings.Contains(*kindStr, "wkfl") {
		*kindStr = "workflow"
	}
}
