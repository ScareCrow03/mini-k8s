package kubeutils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func GetTypeFromYAML(filepath string) string {
	file, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println("read file failed")
	}
	raw := make(map[string]interface{})
	err = yaml.Unmarshal(file, &raw)
	if err != nil {
		fmt.Println("unmarshal yaml failed")
	}
	return raw["kind"].(string)
}
