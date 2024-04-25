package yamlParse

import (
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/logger"
	"os"
)

// 用法：yaml.YAMLParse(&podConfig, "../../assets/pod_config_test1.yaml")，第一个参数传引用
func YAMLParse(ptr interface{}, path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		logger.KError("read yaml failed")
		return err
	}

	err = yaml.Unmarshal(file, ptr)
	if err != nil {
		logger.KError("yaml unmarshal failed")
		return err
	}

	return nil
}
