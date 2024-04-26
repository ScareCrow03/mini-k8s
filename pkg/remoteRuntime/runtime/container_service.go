package rtm

// 这个文件用于提供容器运行时的服务，包括容器的创建、删除、查询等操作；请通过NewRemoteRuntimeService来创建实例，再调用相关的方法；注意它把更底层的镜像服务也包装进了这个对象，外层也可以据此ImgSvc的方法（一般不会用到）；但是Close()方法请注意只能调用外层的！
import (
	"context"
	"encoding/json"
	"io"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	img "mini-k8s/pkg/remoteRuntime/img"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// 描述本文件实现的一些方法，用接口的形式
type ContainerServiceInterface interface {
	CreateContainer(cfg *protocol.ContainerConfig) (string, error)
	StartContainer(containerID string) error
	StopContainer(containerID string) error
	RemoveContainer(containerID string) error
	ListContainers(filterMap ...map[string][]string) ([]types.Container, error)
	InspectContainer(containerID string) (*types.ContainerJSON, error)
	ContainerStatus(containerID string) (*types.StatsJSON, error)
	ExecContainer(containerID string, command []string) (string, error)
	RemoveImageByNameAndItsContainers(imageName string) error
	RemoveImageByIdAndItsContainers(imageId string) error
	RemoveContainerByNameAndItsImage(containerName string, removeImage bool) error
	Close()
}

// 这个服务除了本文件定义的、操作容器的成员方法，还有上述接口声明好的、用于操作Image的方法可以被调用！
type RemoteRuntimeService struct {
	RuntimeClient *client.Client
	ImgSvc        *img.RemoteImageService
}

// 用户只允许通过这里建立一个新的远程运行时服务，底层的ImgService顺便就被建立好了！
func NewRemoteRuntimeService(timeout time.Duration) *RemoteRuntimeService {
	imgSvc := img.NewRemoteImageService(timeout)
	return &RemoteRuntimeService{
		RuntimeClient: imgSvc.ImageClient,
		ImgSvc:        imgSvc,
	}
}

func (r *RemoteRuntimeService) Close() {
	r.ImgSvc.Close()
	r.RuntimeClient = nil
	r.ImgSvc = nil
}

// 不需要pauseId创建的普通方法
func (r *RemoteRuntimeService) CreateContainer(cfg *protocol.ContainerConfig) (string, error) {
	return r.CreateContainerInPod(cfg, "", nil, nil, constant.CtrLabelVal_IsPauseFalse)
}

// 依据protocol中定义的container配置数据结构、以及容器的名字，创建容器，返回长形式的ID
// 本函数支持pauseId的可选指定、如果指定了则加入相应的namespace
// 还支持volumeName->hostPath的转换指定，如果指定了则会挂载相应的卷，否则不会挂载
// 如果有Pod配置，则为它设置标签
func (r *RemoteRuntimeService) CreateContainerInPod(cfg *protocol.ContainerConfig, pauseId string, volumeName2HostPath *map[string]string, pod *protocol.Pod, isPause string) (string, error) {
	// TODO：这个函数需要的参数，涉及到yaml文件的解析，待处理它
	containerConfig, hostConfig, networkConfig, name := cfg.ParseToDockerConfig(volumeName2HostPath, pod, isPause) // 这里传了一个映射指针过去

	// 如果有pause容器的ID，需要进行相关配置，把这些东西都加入到pause的命名空间中
	if pauseId != "" {
		dockerAddToPauseNs := "container:" + pauseId
		hostConfig.NetworkMode = container.NetworkMode(dockerAddToPauseNs)
		hostConfig.IpcMode = container.IpcMode(dockerAddToPauseNs)
		hostConfig.PidMode = container.PidMode(dockerAddToPauseNs)
		// 有了Pause之后，启动这个容器就不需要再设置端口映射了
		containerConfig.ExposedPorts = nil
		hostConfig.Binds = nil
	}

	// 如果有Pod配置
	if pod != nil {
		// 为容器设置标签
		containerConfig.Labels = GetCtrLabelFilterFromPodConfig(&pod.Config, constant.CtrLabelVal_IsPauseFalse)
	}

	err := r.ImgSvc.PullImage(containerConfig.Image, protocol.ImagePullPolicyAtoI(cfg.ImagePullPolicy))
	if err != nil {
		logger.KError("Pull image failed in CreateContainer! Reason: %v", err)
		return "", err
	}

	// 依据上述parse出来适用于docker SDK的配置，创建容器
	resp, err := r.RuntimeClient.ContainerCreate(context.Background(), containerConfig, hostConfig, networkConfig, nil, name)
	if err != nil {
		logger.KError("Create container failed in CreateContainer! Reason: %v", err)
		return "", err
	}

	return resp.ID, nil
}

// 启动容器，如果没有对应的容器，或者提供的是短形式的ID、能对应多个容器，会记录错误信息到日志
func (r *RemoteRuntimeService) StartContainer(containerID string) error {
	err := r.RuntimeClient.ContainerStart(context.Background(), containerID, container.StartOptions{})
	if err != nil {
		logger.KError("Start container failed in StartContainer! Reason: %v", err)
	}
	return err
}

func (r *RemoteRuntimeService) StopContainer(containerID string) error {
	err := r.RuntimeClient.ContainerStop(context.Background(), containerID, container.StopOptions{})
	if err != nil {
		logger.KError("Stop container failed in StopContainer! Reason: %v", err)
	}
	return err
}

// 指定容器ID，删除容器，如果容器正在运行，会先停止它
func (r *RemoteRuntimeService) RemoveContainer(containerID string) error {
	// 尝试获取状态
	containerJSON, err := r.RuntimeClient.ContainerInspect(context.Background(), containerID)
	if err != nil {
		logger.KError("Inspect container failed in RemoveContainer! Reason: %v", err)
		return err
	}

	// 查看状态，如果在运行，先停止
	if containerJSON.State.Running {
		err = r.RuntimeClient.ContainerStop(context.Background(), containerID, container.StopOptions{})
		if err != nil {
			logger.KError("Stop container failed in RemoveContainer! Reason: %v", err)
			return err
		}
	}

	// 删除容器
	err = r.RuntimeClient.ContainerRemove(context.Background(), containerID, container.RemoveOptions{})
	if err != nil {
		logger.KError("Remove container failed in RemoveContainer! Reason: %v", err)
	}
	return err
}

// 列出所有容器（可以在参数中指定key-val过滤条件，应该是若干个 1个key到若干个val的映射），返回docker SDK定义的容器信息列表
// 如果未指定过滤器/指定为空，会列出所有容器；否则根据列出的key-val条件，同一个key的多个value求并集，不同的key求交集做筛选
// 注意，Pod层面的Selector按照Label的筛选逻辑是在更上层实现的！Container无需关心
func (r *RemoteRuntimeService) ListContainers(filterMap ...map[string][]string) ([]types.Container, error) {
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

	// 获取过滤后的容器列表
	containers, err := r.RuntimeClient.ContainerList(context.Background(), container.ListOptions{All: true, Filters: filterArgs})
	if err != nil {
		logger.KError("List containers failed in ListContainers! Reason: %v", err)
		return nil, err
	}
	return containers, nil
}

// 此函数返回容器的详细信息，包括容器的配置、网络设置、挂载点、状态等。这个函数主要用于获取容器的配置和状态信息。
func (r *RemoteRuntimeService) InspectContainer(containerID string) (*types.ContainerJSON, error) {
	containerJSON, err := r.RuntimeClient.ContainerInspect(context.Background(), containerID)
	if err != nil {
		logger.KError("Inspect container failed in InspectContainer! Reason: %v", err)
	}
	return &containerJSON, err
}

// 返回给定容器的近实时统计信息。这些统计信息包括CPU使用率、内存使用情况、网络使用情况等。这个函数主要用于监控和分析容器的运行性能。
func (r *RemoteRuntimeService) ContainerStatus(containerID string) (*types.StatsJSON, error) {

	stats, err := r.RuntimeClient.ContainerStats(context.Background(), containerID, false)
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
func (r *RemoteRuntimeService) ExecContainer(containerID string, command []string) (string, error) {
	logger.KInfo("Exec container %s with command %v", containerID, command)
	// 创建一个exec配置
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	}

	// 在容器中创建一个exec实例
	execID, err := r.RuntimeClient.ContainerExecCreate(context.Background(), containerID, execConfig)
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
	resp, err := r.RuntimeClient.ContainerExecAttach(context.Background(), execID.ID, execStartCheck)
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

// 指定镜像名，删除这个镜像及其所有容器；谨慎使用！同一个docker守护进程管理的容器名确实是唯一的，
// k8s中，一个Pod里的容器名必须唯一、否则报错；但不同的Pod里的容器名可以重复；k8s通常使用Pod名与Pod之下的容器名来定位容器，上述都指的是在某个Pod的context下的容器名字，而不是本机docker管理的全局容器名
func (r *RemoteRuntimeService) RemoveImageByNameAndItsContainers(imageName string) error {
	summary, err := r.ImgSvc.GetImage(imageName)
	if err != nil {
		logger.KError("Get image failed in RemoveImageByNameAndItsContainer! Reason: %v", err)
		return err
	}

	if summary.ID == "" {
		logger.KInfo("Image %s not found, no need remove", imageName)
		return nil
	}

	return r.RemoveImageByIdAndItsContainers(summary.ID)
}

func (r *RemoteRuntimeService) RemoveImageByIdAndItsContainers(imageId string) error {
	logger.KInfo("Removing imageId %s and its containers", imageId)

	// 获取指定镜像的信息；这其实是在检验这个镜像是否存在
	imageInfo, _, err := r.RuntimeClient.ImageInspectWithRaw(context.Background(), imageId)
	if err != nil { // 也许这里并不能算一个错误，因为找不到镜像也是正常情况、那么此时无需删除即可
		logger.KWarning("Failed to inspect imageId %s,  maybe no need to remove! Reason: %v", imageId, err)
		return err
	}

	// 获取所有的容器，否则只会获取运行中的容器
	containerList, err := r.RuntimeClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		logger.KError("Failed to list containers! Reason: %v", err)
		return err
	}

	// 遍历所有的容器
	for _, containerInfo := range containerList {
		// 如果容器使用的是指定的镜像；应该比较ID！
		if containerInfo.ImageID == imageInfo.ID {
			// 停止容器
			if err := r.RuntimeClient.ContainerStop(context.Background(), containerInfo.ID, container.StopOptions{}); err != nil {
				logger.KError("Failed to stop container %s! Reason: %v", containerInfo.ID, err)
				return err
			}

			// 删除容器
			if err := r.RuntimeClient.ContainerRemove(context.Background(), containerInfo.ID, container.RemoveOptions{}); err != nil {
				logger.KError("Failed to remove container %s! Reason: %v", containerInfo.ID, err)
				return err
			}

			logger.KInfo("Successfully removed container %s", containerInfo.ID)
		}
	}

	// 删除镜像
	_, err = r.RuntimeClient.ImageRemove(context.Background(), imageInfo.ID, image.RemoveOptions{Force: true})
	if err != nil {
		logger.KError("Failed to remove image %s! Reason: %v", imageId, err)
		return err
	}

	logger.KInfo("Successfully removed image %s", imageId)
	return nil
}

// 指定docker管理的全局容器名，删除这个容器、可选删除它的镜像；这个方法一般不会被用到，因为如果多个Pod里、指定相同的Pod.ContainerName，那么这些名字被实际应用到docker上时、会加入一些后缀以保证docker守护进程管理的唯一性，则此时没有上下文地指定容器名（只能理解为docker管理的直接容器名）是不合理的
func (r *RemoteRuntimeService) RemoveContainerByNameAndItsImage(containerName string, removeImage bool) error {
	logger.KInfo("Removing container By Name %s, need to remote Its Image: %v", containerName, removeImage)

	// 获取所有的容器
	containerList, err := r.RuntimeClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		logger.KError("Failed to list containers! Reason: %v", err)
		return err
	}

	// 遍历所有的容器
	for _, containerInfo := range containerList {
		// 如果找到了指定的容器
		//在所有的容器中查找名字与指定名字相同的容器。在Docker中，容器的名字在Names字段中，这是一个字符串数组，因为一个容器可能有多个名字。这里我们只查看第一个名字（Names[0]），并且由于Docker的名字前面会有一个"/“，所以我们在比较时也加上了”/"。
		if containerInfo.Names[0] == "/"+containerName {
			// 停止容器
			if err := r.RuntimeClient.ContainerStop(context.Background(), containerInfo.ID, container.StopOptions{}); err != nil {
				logger.KError("Failed to stop container %s! Reason: %v", containerInfo.ID, err)
				return err
			}

			// 删除容器
			if err := r.RuntimeClient.ContainerRemove(context.Background(), containerInfo.ID, container.RemoveOptions{}); err != nil {
				logger.KError("Failed to remove container %s! Reason: %v", containerInfo.ID, err)
				return err
			}

			logger.KInfo("Successfully removed container %s", containerInfo.ID)

			// 如果需要删除镜像
			if removeImage {
				// 删除镜像
				_, err = r.RuntimeClient.ImageRemove(context.Background(), containerInfo.ImageID, image.RemoveOptions{Force: true})
				if err != nil {
					logger.KError("Failed to remove image %s! Reason: %v", containerInfo.ImageID, err)
					return err
				}

				logger.KInfo("Successfully removed image %s", containerInfo.ImageID)
			}

			// 返回成功
			return nil
		}
	}

	// 如果没有找到指定的容器
	logger.KInfo("Container %s not found", containerName)
	return nil
}
