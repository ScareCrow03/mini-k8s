package remoteRuntime

// 这个文件用于提供容器运行时的服务，包括容器的创建、删除、查询等操作；请通过NewRemoteRuntimeService来创建实例，再调用相关的方法；注意它把更底层的镜像服务也包装进了这个对象，外层也可以据此ImgSvc的方法（一般不会用到）；但是Close()方法请注意只能调用外层的！
import (
	"context"
	"encoding/json"
	"io"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type RuntimeServiceInterface interface { // 描述本文件实现了哪些方法，仅做声明用（视为一个本文件的摘要）；可以利用remoteRuntimeService的成员方法来调用它们
	CreateContainer(cfg *protocol.ContainerConfig) (string, error)
	StartContainer(containerID string) error
	StopContainer(containerID string) error
	RemoveContainer(containerID string) error
	ListContainers(filterMap ...map[string][]string) ([]types.Container, error)
	InspectContainer(containerID string) (*types.ContainerJSON, error)
	ContainerStatus(containerID string) (*types.StatsJSON, error)
	ExecContainer(containerID string, command []string) (string, error)
	Close()
	// TODO: 待实现其他的方法
}

// 这个服务除了本文件定义的、操作容器的成员方法，还有上述接口声明好的、用于操作Image的方法可以被调用！
type remoteRuntimeService struct {
	runtimeClient *client.Client
	ImgSvc        *remoteImageService
}

// 用户只允许通过这里建立一个新的远程运行时服务，底层的ImgService顺便就被建立好了！
func NewRemoteRuntimeService(timeout time.Duration) *remoteRuntimeService {
	imgSvc := newRemoteImageService(timeout)
	return &remoteRuntimeService{
		runtimeClient: imgSvc.imageClient,
		ImgSvc:        imgSvc,
	}
}

func (r *remoteRuntimeService) Close() {
	r.ImgSvc.Close()
	r.runtimeClient = nil
	r.ImgSvc = nil
}

// 依据protocol中定义的container配置数据结构，创建容器，返回长形式的ID
func (r *remoteRuntimeService) CreateContainer(cfg *protocol.ContainerConfig, alwaysPull bool) (string, error) {
	// TODO：这个函数需要的参数，涉及到yaml文件的解析，待处理它
	containerConfig, hostConfig, name := cfg.ParseToDockerConfig()

	err := r.ImgSvc.PullImage(containerConfig.Image, alwaysPull)
	if err != nil {
		logger.KError("Pull image failed in CreateContainer! Reason: %v", err)
		return "", err
	}

	// 依据上述parse出来适用于docker SDK的配置，创建容器
	resp, err := r.runtimeClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, name)
	if err != nil {
		logger.KError("Create container failed in CreateContainer! Reason: %v", err)
		return "", err
	}

	return resp.ID, nil
}

// 启动容器，如果没有对应的容器，或者提供的是短形式的ID、能对应多个容器，会记录错误信息到日志
func (r *remoteRuntimeService) StartContainer(containerID string) error {
	err := r.runtimeClient.ContainerStart(context.Background(), containerID, container.StartOptions{})
	if err != nil {
		logger.KError("Start container failed in StartContainer! Reason: %v", err)
	}
	return err
}

func (r *remoteRuntimeService) StopContainer(containerID string) error {
	err := r.runtimeClient.ContainerStop(context.Background(), containerID, container.StopOptions{})
	if err != nil {
		logger.KError("Stop container failed in StopContainer! Reason: %v", err)
	}
	return err
}

// 删除容器，如果容器正在运行，会先停止它
func (r *remoteRuntimeService) RemoveContainer(containerID string) error {
	// 尝试获取状态
	containerJSON, err := r.runtimeClient.ContainerInspect(context.Background(), containerID)
	if err != nil {
		logger.KError("Inspect container failed in RemoveContainer! Reason: %v", err)
		return err
	}

	// 查看状态，如果在运行，先停止
	if containerJSON.State.Running {
		err = r.runtimeClient.ContainerStop(context.Background(), containerID, container.StopOptions{})
		if err != nil {
			logger.KError("Stop container failed in RemoveContainer! Reason: %v", err)
			return err
		}
	}

	// 删除容器
	err = r.runtimeClient.ContainerRemove(context.Background(), containerID, container.RemoveOptions{})
	if err != nil {
		logger.KError("Remove container failed in RemoveContainer! Reason: %v", err)
	}
	return err
}

// 列出所有容器（可以在参数中指定key-val过滤条件，应该是若干个 1个key到若干个val的映射），返回docker SDK定义的容器信息列表
// 如果未指定过滤器/指定为空，会列出所有容器；否则根据列出的key-val条件，同一个key的多个value求并集，不同的key求交集做筛选
// 注意，Pod层面的Selector按照Label的筛选逻辑是在更上层实现的！Container无需关心
func (r *remoteRuntimeService) ListContainers(filterMap ...map[string][]string) ([]types.Container, error) {
	// 将go语言内置的key-val映射，对接到docker SDK提供的过滤器
	var filterArgs filters.Args
	if len(filterMap) > 0 && filterMap[0] != nil {
		filterArgs = filters.NewArgs()
		for k, v := range filterMap[0] {
			for _, val := range v {
				filterArgs.Add(k, val)
			}
		}
	}

	// 先拿到所有容器的列表
	containers, err := r.runtimeClient.ContainerList(context.Background(), container.ListOptions{Filters: filterArgs})
	if err != nil {
		logger.KError("List containers failed in ListContainers! Reason: %v", err)
		return nil, err
	}
	return containers, nil
}

// 此函数返回容器的详细信息，包括容器的配置、网络设置、挂载点、状态等。这个函数主要用于获取容器的配置和状态信息。
func (r *remoteRuntimeService) InspectContainer(containerID string) (*types.ContainerJSON, error) {
	containerJSON, err := r.runtimeClient.ContainerInspect(context.Background(), containerID)
	if err != nil {
		logger.KError("Inspect container failed in InspectContainer! Reason: %v", err)
	}
	return &containerJSON, err
}

// 返回给定容器的近实时统计信息。这些统计信息包括CPU使用率、内存使用情况、网络使用情况等。这个函数主要用于监控和分析容器的运行性能。
func (r *remoteRuntimeService) ContainerStatus(containerID string) (*types.StatsJSON, error) {

	stats, err := r.runtimeClient.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		logger.KError("Get container status failed in ContainerStatus! Reason: %v", err)
		return nil, err
	}
	defer stats.Body.Close()

	// 读取统计信息的内容
	data, err := io.ReadAll(stats.Body)
	if err != nil {
		logger.KError("Read container status failed in ContainerStatus! Reason: %v", err)
		return nil, err
	}

	// 解析统计信息为*types.StatsJSON
	var statsJSON types.StatsJSON
	err = json.Unmarshal(data, &statsJSON)
	if err != nil {
		logger.KError("Unmarshal container status to types.StatsJSON failed in ContainerStatus! Reason: %v", err)
		return nil, err
	}

	// 返回容器的统计信息
	return &statsJSON, nil
}

// 在容器内运行一段cmd命令，接收stdout和stderr的输出并返回
func (r *remoteRuntimeService) ExecContainer(containerID string, command []string) (string, error) {
	logger.KInfo("Exec container %s with command %v", containerID, command)
	// 创建一个exec配置
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	}

	// 在容器中创建一个exec实例
	execID, err := r.runtimeClient.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		logger.KError("Create exec instance failed in ExecContainer! Reason: %v", err)
		return "", err
	}

	// 创建一个ExecStartCheck配置
	execStartCheck := types.ExecStartCheck{
		Detach: false,
		Tty:    false,
	}

	// 运行这个exec实例
	resp, err := r.runtimeClient.ContainerExecAttach(context.Background(), execID.ID, execStartCheck)
	if err != nil {
		logger.KError("Exec attach failed in ExecContainer! Reason: %v", err)
		return "", err
	}
	defer resp.Close()

	// 读取并返回执行结果
	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		logger.KError("Read exec output failed in ExecContainer! Reason: %v", err)
		return "", err
	}

	return string(output), nil
}

func (r *remoteRuntimeService) RemoveImageAndContainers(imageName string) error {
	logger.KInfo("Remove image %s and containers using it", imageName)
	// 创建一个过滤器来选择使用指定镜像的容器
	filterArgs := filters.NewArgs()
	filterArgs.Add("ancestor", imageName)

	// 获取使用指定镜像的所有容器
	containers, err := r.runtimeClient.ContainerList(context.Background(), container.ListOptions{Filters: filterArgs})
	if err != nil {
		return err
	}

	// 删除这些容器
	for _, c := range containers {
		err = r.runtimeClient.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{Force: true})
		if err != nil {
			return err
		}
	}

	// 删除指定的镜像
	_, err = r.runtimeClient.ImageRemove(context.Background(), imageName, image.RemoveOptions{Force: true})
	if err != nil {
		return err
	}

	return nil
}
