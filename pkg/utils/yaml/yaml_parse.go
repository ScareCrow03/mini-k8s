package yamlParse

import (
	"os"

	"gopkg.in/yaml.v3"
)

// 用法：yaml.YAMLParse(&podConfig, "../../assets/pod_config_test1.yaml")，第一个参数传引用
func YAMLParse(ptr interface{}, path string) error {
	// logger.KDebug(path)
	file, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(file, ptr)
	if err != nil {
		panic(err)
	}

	return nil
}
