// env_test.go
package env

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestMasterEnvLoadConfig(t *testing.T) {
	// 测试加载master 节点的配置文件
	err := MstEnvCfg.LoadConfig("../../../cluster_env_config/master_env_config.yml")
	if err != nil {
		t.Errorf("Failed to load master config: %v", err)
	}

	// 检查加载的配置是否正确
	if MstEnvCfg.Etcd.Address == "" || MstEnvCfg.Etcd.Timeout < 5 ||
		MstEnvCfg.Kafka.Address == "" || MstEnvCfg.MiniK8s.ApiServerAddress == "" {
		t.Errorf("Master config was not loaded correctly")
	}

	// 将配置数据结构转换为 YAML 格式
	data, err := yaml.Marshal(&MstEnvCfg)
	if err != nil {
		t.Errorf("Failed to marshal config: %v", err)
	}

	// 在命令行中打印配置数据结构
	fmt.Printf("--- Master Config:\n%s\n\n", string(data))
}

func TestWorkerEnvLoadConfig(t *testing.T) {
	// 测试加载worker节点的配置文件
	err := WrkEnvCfg.LoadConfig("../../../cluster_env_config/worker_env_config.yml")
	if err != nil {
		t.Errorf("Failed to load worker config: %v", err)
	}

	// 检查加载的配置是否正确
	if WrkEnvCfg.Kafka.Address == "" || WrkEnvCfg.MiniK8s.ApiServerAddress == "" {
		t.Errorf("Worker config was not loaded correctly")
	}

	// 将配置数据结构转换为 YAML 格式
	data, err := yaml.Marshal(&WrkEnvCfg)
	if err != nil {
		t.Errorf("Failed to marshal config: %v", err)
	}

	// 在命令行中打印配置数据结构
	fmt.Printf("--- Worker Config:\n%s\n\n", string(data))
}
