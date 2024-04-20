package protocol

// 这个文件用于自行设计简化后的.yml文件、在指定container配置时需要解析哪些字段，然后放在ContainerConfig里，后续它被parse为能直接传给docker SDK的Config和HostConfig
// 如果需要解析yaml文件中的更多字段，可以在这里添加，然后在parse方法里添加解析逻辑

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

// 拉取镜像的三种策略，建立一个枚举类型
type ImagePullPolicyType int

const (
	AlwaysPull ImagePullPolicyType = iota // 从0开始枚举
	PullIfNotPresent
	NeverPull
)

// 用户启动这个k8s集群时、它希望创建的容器应该与普通的容器区分开，考虑设计相关独特的key-val标志？
type ContainerConfig struct {
	Name  string `yaml:"name"`  // container name
	Image string `yaml:"image"` // container image name and version，like "nginx:latest"

	Command []string `yaml:"command"` // 容器启动命令列表
	Args    []string `yaml:"args"`    // 它与command字段一起构成容器启动命令

	ImagePullPolicy string `yaml:"imagePullPolicy"` // 这个字段在CreateContainer方法中被使用一次，不直接参与docker SDK的配置
	WorkingDir      string `yaml:"workingDir"`
	Ports           []struct {
		Name          string `yaml:"name"` // 只用于某个标识端口映射关系的名字、用于向用户提示它的意义，docker SDK不会使用它
		ContainerPort int64  `yaml:"containerPort"`
		HostPort      int64  `yaml:"hostPort"`
		Protocol      string `yaml:"protocol"` // 允许为空，此时默认为"tcp"，否则按道理只能为"tcp"或"udp"。此处放宽限制，如果这个字段不指定为udp，那么它都是tcp
	} `yaml:"ports"`
	VolumeMounts []struct {
		Name      string `yaml:"name"`
		MountPath string `yaml:"mountPath"`
		ReadOnly  bool   `yaml:"readOnly"`
	} `yaml:"volumeMounts"`

	Resources struct {
		Limits struct {
			cpu    string `yaml:"cpu"`
			memory string `yaml:"memory"`
		}
	}
	Env map[string]string `yaml:"env"`
}

// 将字符串转换为ImagePullPolicy枚举类型；默认策略为PullIfNotPresent
func ImagePullPolicyAtoI(policy string) ImagePullPolicyType {
	switch policy {
	case "Always":
		return AlwaysPull
	case "IfNotPresent":
		return PullIfNotPresent
	case "Never":
		return NeverPull
	default:
		return AlwaysPull
	}
}

// TODO: 从上述ContainerConfig解析出docker SDK需要的Config, HostConfig, 容器名字符串；其中某些字段可能在docker config中没有对应的字段，而是作为一些额外的信息提供给runtime方法
func (c *ContainerConfig) ParseToDockerConfig() (*container.Config, *container.HostConfig, *network.NetworkingConfig, string) {
	config := &container.Config{
		Image:      c.Image,
		Cmd:        append(c.Command, c.Args...),
		WorkingDir: c.WorkingDir,
		Env:        convertMapToSlice(c.Env),
	}

	// 创建 container.HostConfig
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{},
		Binds:        []string{},
		Resources: container.Resources{
			CPUShares: parseCPUShares(c.Resources.Limits.cpu),
			Memory:    parseMemory(c.Resources.Limits.memory),
		},
	}

	// 暴露端口相关的配置，格式为"containerPort:HostPort"
	for _, port := range c.Ports {
		protocol := port.Protocol
		if protocol != "udp" {
			protocol = "tcp"
		}
		hostConfig.PortBindings[nat.Port(fmt.Sprintf("%d/tcp", port.ContainerPort))] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: fmt.Sprintf("%d", port.HostPort),
			},
		}
	}

	// 挂载卷相关的配置，格式为"volumeName:containerPath:ro/rw"
	for _, volume := range c.VolumeMounts {
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s:%s", volume.Name, volume.MountPath, getMountMode(volume.ReadOnly)))
	}

	// 创建 network.NetworkingConfig，默认为空
	networkingConfig := &network.NetworkingConfig{}

	return config, hostConfig, networkingConfig, c.Name
}

func convertMapToSlice(envMap map[string]string) []string {
	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func getMountMode(readOnly bool) string {
	if readOnly {
		return "ro"
	}
	return "rw"
}

func parseCPUShares(cpu string) int64 {
	// 这里假设 cpu 的单位是 core 数，1 core 对应 1024 shares
	shares, _ := strconv.ParseInt(cpu, 10, 64)
	return shares * 1024
}

func parseMemory(memory string) int64 {
	// 这里假设 memory 的单位是 MiB 或 GiB
	if strings.HasSuffix(memory, "GiB") {
		memory = strings.TrimSuffix(memory, "GiB")
		mib, _ := strconv.ParseInt(memory, 10, 64)
		return mib * 1024 * 1024 * 1024
	} else if strings.HasSuffix(memory, "MiB") {
		memory = strings.TrimSuffix(memory, "MiB")
		mib, _ := strconv.ParseInt(memory, 10, 64)
		return mib * 1024 * 1024
	}
	return 0
}
