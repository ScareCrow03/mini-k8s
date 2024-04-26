package env

// 这个文件提供了一个全局变量MstCfgEnv，用于存储master节点的环境配置信息；它在init函数中初始化了一些默认值，然后可以在main函数中通过调用LoadConfig函数来加载用户设定的配置文件；在此之后，可以通过引用这个包来访问到这个全局变量的一些字段信息。
// 与初始化有关的东西，请不要写入到logger中，防止递归引用包
import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	MstEnvCfg        MasterEnvConfigType // 可以在初始化时被修改，外部只读
	defaultMsEnvtCfg MasterEnvConfigType // 只被初始化一次，外部不可见
)

type MasterEnvConfigType struct {
	Etcd struct {
		Address string `yaml:"address"`
		Timeout int    `yaml:"timeout"` // 单位秒
	} `yaml:"etcd"`
	Kafka struct {
		Address string `yaml:"address"`
	} `yaml:"kafka"`
	MiniK8s struct {
		ApiServerAddress string `yaml:"apiServerAddress"`
	} `yaml:"mini-k8s"`
}

// 默认值
func init() {
	defaultMsEnvtCfg.SetDefault()

	MstEnvCfg = defaultMsEnvtCfg
}

func (cfg *MasterEnvConfigType) SetDefault() {
	cfg.Etcd.Address = "localhost:2379"
	cfg.Etcd.Timeout = 5
	cfg.Kafka.Address = "localhost:9092"
	cfg.MiniK8s.ApiServerAddress = "localhost:6443"
}

// 为了开发方便，每个Master进程只按以下方法读取一遍配置文件（如果它的main函数不读取，那么采用上述的默认值），如果后续发生连接错误，立即退出进程！
func (cfg *MasterEnvConfigType) LoadConfig(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("read user_master_config yaml failed, maybe no need to loadConfig from file! Reason: %s\n", err)
		return err
	}

	// 如果没有检测到某个字段，那么会保留原始值
	err = yaml.Unmarshal(file, cfg)
	if err != nil {
		fmt.Printf("read user_master_config yaml failed, Reason: %s\n", err)
		return err
	}

	if cfg.Etcd.Address == "" {
		cfg.Etcd.Address = defaultMsEnvtCfg.Etcd.Address
	}

	if cfg.Etcd.Timeout < 5 {
		cfg.Etcd.Timeout = defaultMsEnvtCfg.Etcd.Timeout
	}

	if cfg.Kafka.Address == "" {
		cfg.Kafka.Address = defaultMsEnvtCfg.Kafka.Address
	}

	if cfg.MiniK8s.ApiServerAddress == "" {
		cfg.MiniK8s.ApiServerAddress = defaultMsEnvtCfg.MiniK8s.ApiServerAddress
	}

	return nil
}
