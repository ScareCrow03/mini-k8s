package protocol

// 这个文件用于自行设计简化后的.yml文件、在指定container配置时需要解析哪些字段，然后放在ContainerConfig里，后续它被parse为能直接传给docker SDK的Config和HostConfig
// 如果需要解析yaml文件中的更多字段，可以在这里添加，然后在parse方法里添加解析逻辑

import (
	"fmt"
	"mini-k8s/pkg/constant"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"gopkg.in/yaml.v3"
)

// 拉取镜像的三种策略，建立一个枚举类型
type ImagePullPolicyType int

const (
	AlwaysPull ImagePullPolicyType = iota // 从0开始枚举
	PullIfNotPresent
	NeverPull

	AlwaysPullStr       = "Always"
	PullIfNotPresentStr = "IfNotPresent"
	NeverPullStr        = "Never"
)

// 用户启动这个k8s集群时、它希望创建的容器应该与普通的容器区分开，考虑设计相关独特的key-val标志？
type ContainerConfig struct {
	UID             string               `yaml:"uid" json:"uid"`                         // 这个字段是在容器创建时生成的，然后需要回填进Pod相关的结构体内，使得本Pod能够追踪到自己的容器。是否需要这个字段？
	Name            string               `yaml:"name" json:"name"`                       // container name
	Image           string               `yaml:"image" json:"image"`                     // container image name and version，like "nginx:latest"
	Labels          map[string]string    `yaml:"labels" json:"labels"`                   // 容器的label字段用户不需要指定，仅用于底层操作容器方便
	Command         []string             `yaml:"command" json:"command"`                 // 容器启动命令列表
	Args            []string             `yaml:"args" json:"args"`                       // 它与command字段一起构成容器启动命令
	ImagePullPolicy string               `yaml:"imagePullPolicy" json:"imagePullPolicy"` // 这个字段在CreateContainer方法中被使用一次，不直接参与docker SDK的配置
	WorkingDir      string               `yaml:"workingDir" json:"workingDir"`
	Ports           []CtrPortBindingType `yaml:"ports" json:"ports"`
	VolumeMounts    []CtrVolumeMountType `yaml:"volumeMounts" json:"volumeMounts"` // 这个类型与Pod配置紧耦合，因为它需要Pod提供的、某个VolumeName对应的hostPath
	Resources       CtrResourcesType     `yaml:"resources" json:"resources"`
	Env             map[string]string    `yaml:"env" json:"env"`
}

type CtrResourcesType struct {
	Limits struct {
		CPU    string `yaml:"cpu" json:"cpu"`
		Memory string `yaml:"memory" json:"memory"`
	}
}

type CtrPortBindingType struct {
	Name          string `yaml:"name" json:"name"` // 只用于某个标识端口映射关系的名字、用于向用户提示它的意义，docker SDK不会使用它
	ContainerPort int64  `yaml:"containerPort" json:"containerPort"`
	HostPort      int64  `yaml:"hostPort" json:"hostPort"`
	Protocol      string `yaml:"protocol" json:"protocol"` // 允许为空，此时默认为"tcp"，否则按道理只能为"tcp"或"udp"。此处放宽限制，如果这个字段不指定为udp，那么它都是tcp
}

type CtrVolumeMountType struct {
	Name      string `yaml:"name" json:"name"`
	MountPath string `yaml:"mountPath" json:"mountPath"`
	ReadOnly  bool   `yaml:"readOnly" json:"readOnly"`
}

// 将字符串转换为ImagePullPolicy枚举类型；默认策略为PullIfNotPresent
func ImagePullPolicyAtoI(policy string) ImagePullPolicyType {
	switch policy {
	case AlwaysPullStr:
		return AlwaysPull
	case PullIfNotPresentStr:
		return PullIfNotPresent
	case NeverPullStr:
		return NeverPull
	default:
		return PullIfNotPresent
	}
}

// TODO: 从上述ContainerConfig解析出docker SDK需要的Config, HostConfig, 容器名字符串；其中某些字段可能在docker config中没有对应的字段，而是作为一些额外的信息提供给runtime方法
func (c *ContainerConfig) ParseToDockerConfig(volumeName2HostPath *map[string]string, pod *Pod, isPause string) (*container.Config, *container.HostConfig, *network.NetworkingConfig, string) {
	config := &container.Config{
		Image:        c.Image,
		Cmd:          append(c.Command, c.Args...),
		WorkingDir:   c.WorkingDir,
		Env:          convertMapToSlice(c.Env),
		ExposedPorts: map[nat.Port]struct{}{},
		Labels:       c.Labels,
	}

	// 创建 container.HostConfig
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{},
		Binds:        []string{},
		Resources: container.Resources{
			CPUShares: parseCPUShares(c.Resources.Limits.CPU),
			Memory:    parseMemory(c.Resources.Limits.Memory),
		},
	}

	if isPause == constant.CtrLabelVal_IsPauseTrue {
		// 如果是pause容器，为它设置IPCMode为Shareable
		hostConfig.IpcMode = container.IPCModeShareable
		// 以及，分配一个flannel的IP；这是在将这个容器加入预设好的flannel网络，需要保证相关环境已经配置正确！
		hostConfig.NetworkMode = container.NetworkMode("flannel")
	}

	// 暴露端口相关的配置，其实是建立一个containerPort->hostIP,hostPort的映射
	if pod == nil {
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
	}

	// 挂载卷相关的配置，格式为"hostPath:containerPath:ro/rw"
	// 如果不指定name到hostPath的map，那么不做挂载；对于从映射关系中找不到name的volume，也不做挂载
	if volumeName2HostPath != nil {
		for _, volume := range c.VolumeMounts {
			data, _ := yaml.Marshal(volume)
			fmt.Println(string(data))
			mappingHostPath := (*volumeName2HostPath)[volume.Name]
			fmt.Println(mappingHostPath)
			if mappingHostPath == "" {
				continue
			}
			hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s:%s", mappingHostPath, volume.MountPath, getMountMode(volume.ReadOnly)))
			data, _ = yaml.Marshal(hostConfig.Binds)
			fmt.Println(string(data))
		}
	}
	// 虽然可以设定hostConfig的RestartPolicy，但是很麻烦，需要结合Pod的生命周期来做，所以这里不做设置

	// 创建 network.NetworkingConfig，默认为空
	networkingConfig := &network.NetworkingConfig{}

	// 如果指定了pod，那么容器名为"minik8s-container名+podId"，否则为container名
	ctrName := c.Name
	if pod != nil {
		ctrName = fmt.Sprintf("minik8s-%s-%s", c.Name, pod.Config.Metadata.UID)
	}
	return config, hostConfig, networkingConfig, ctrName
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
