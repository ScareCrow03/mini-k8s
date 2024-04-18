package protocol

// 这个文件用于自行设计简化后的.yml文件、在指定container配置时需要解析哪些字段，然后放在ContainerConfig里，后续它被parse为能直接传给docker SDK的Config和HostConfig
// 如果需要解析yaml文件中的更多字段，可以在这里添加，然后在parse方法里添加解析逻辑

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
)

// 用户启动这个k8s集群时、它希望创建的容器应该与普通的容器区分开，考虑设计相关独特的key-val标志？
type ContainerConfig struct {
	// 必须指定的字段
	Image string // container image name and version，like "nginx:latest"

	// 可选指定的字段；若未指定则采用0默认值、或者docker随机生成的值
	Name         string                // container name，默认随机生成
	Cmd          []string              // container command，未指定则采用默认值
	CPUShares    int64                 // CPU usage，未指定则不作限制
	Memory       int64                 // Memory limit in bytes，未指定则不作限制
	Binds        []string              // volume mounts，未指定则不挂载额外空间
	ExposedPorts map[nat.Port]struct{} // exposed ports，未指定则不暴露端口
	// 待添加更多字段
}

// TODO: 从上述ContainerConfig解析出docker SDK需要的Config, HostConfig, 容器名字符串
func (c *ContainerConfig) ParseToDockerConfig() (*container.Config, *container.HostConfig, string) {
	containerConfig := &container.Config{
		Image:        c.Image,
		Cmd:          strslice.StrSlice(c.Cmd),
		ExposedPorts: c.ExposedPorts,
	}
	hostConfig := &container.HostConfig{
		Binds: c.Binds,
		Resources: container.Resources{
			CPUShares: c.CPUShares,
			Memory:    c.Memory,
		},
	}
	return containerConfig, hostConfig, c.Name
}
