package rtm

import (
	"errors"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	flannelClient "mini-k8s/pkg/utils/cni/flannel"

	"github.com/docker/docker/api/types"
)

// 这个文件实现了CRI标准定义的，PodSandbox级别的操作，除了RunSandbox只会运行pause外，停止、删除Sandbox方法均对于本pod所有容器生效；
// 这个文件是对RemoteRuntimeService的其他方法实现，概念提高到Pod层面
type PodSandboxServiceInterface interface {
	RunPodSandBox(pod *protocol.Pod) (string, error)
	StopPodSandBox(pod *protocol.Pod) error
	RemovePodSandBox(pod *protocol.Pod) error
	PodSandBoxStatus(pod *protocol.Pod) (*types.ContainerJSON, error)
	ListPodSandBox(filter map[string][]string) ([]types.Container, error)
}

func GetPauseConfigFromPodConfig(podConfig *protocol.PodConfig) *protocol.ContainerConfig {
	// 从PodConfig中提取pause容器需要的配置，主要是同一个namespce的端口映射配置、必须在pause启动前准备好
	pauseConfigRes := protocol.ContainerConfig{
		Name:            constant.CtrPauseName_Prefix + podConfig.Metadata.UID,
		Image:           constant.CtrPauseImgUrl,
		ImagePullPolicy: protocol.PullIfNotPresentStr,
	}

	// 从PodConfig中提取端口映射配置，行为是访问它包含的每一个container的端口映射配置，全部添加到pause容器的配置中
	for _, containerConfig := range podConfig.Spec.Containers {
		pauseConfigRes.Ports = append(pauseConfigRes.Ports, containerConfig.Ports...)
	}

	// TODO: 可能需要添加一些label，方便Pod能找到这个pause容器；最基本地，标注本容器归属于这个Pod，以及这是一个pause容器
	pauseConfigRes.Labels = GetCtrLabelFilterFromPodConfig(podConfig, constant.CtrLabelVal_IsPauseTrue)
	// TODO END
	return &pauseConfigRes
}

// 从Pod中提取容器的label，以便能够找到这些容器，它是一种通用的查找方法，对于本Pod的其他容器也适用
func GetCtrLabelFilterFromPodConfig(podConfig *protocol.PodConfig, isPause string) map[string]string {
	res := make(map[string]string)
	// 按需添加key-value
	if podConfig.Metadata.UID != "" {
		res[constant.CtrLabel_PodId] = podConfig.Metadata.UID
	}
	if podConfig.Metadata.Name != "" {
		res[constant.CtrLabel_PodName] = podConfig.Metadata.Name
	}
	if podConfig.Metadata.Namespace != "" {
		res[constant.CtrLabel_PodNamespace] = podConfig.Metadata.Namespace
	}
	if isPause == constant.CtrLabelVal_IsPauseFalse || isPause == constant.CtrLabelVal_IsPauseTrue {
		res[constant.CtrLabel_IsPause] = isPause
	}
	return res
}

// 创建并启动一个pause容器；如果多次调用，那么每次都会先把前一个pause容器停掉、移除，然后在干净的环境下创建新的pause容器
// 请注意，Pod的UID应该在api-server把配置写入etcd时就生成，而不是在kubelet真正开始创建时生成
func (r *RemoteRuntimeService) RunPodSandBox(pod *protocol.Pod) (string, error) {
	// 清理掉旧的sandbox，保持一个干净的环境
	r.RemovePodSandBox(pod)

	// 建立一个引用
	podConfig := &pod.Config

	// 根据Pod的配置，创建pause容器的配置；这包含做一些端口映射的整合，以及
	pauseConfig := GetPauseConfigFromPodConfig(podConfig)

	// 创建pause容器
	pauseContainerID, err := r.CreateContainerInPod(pauseConfig, "", nil, nil, constant.CtrLabelVal_IsPauseTrue)
	if err != nil {
		logger.KError("Failed to create pause container: %v", err)
		return "", err
	}
	// 启动pause容器
	err = r.StartContainer(pauseContainerID)
	if err != nil {
		logger.KError("Failed to start pause container: %v", err)
		return "", err
	}

	// 将pause容器加入flannel网络，即得到PodIP；这个Pod如果重启，那么分配到的新PodIP可以不一样、是动态的
	podIp, err := flannelClient.LookupIP(pauseContainerID)
	if err != nil {
		logger.KError("Failed to lookup pod IP: %v", err)
	}
	// 写回
	pod.Status = protocol.PodStatus{
		Phase: constant.PodPhasePending,
		IP:    podIp,
	}
	return pauseContainerID, nil
}

func (r *RemoteRuntimeService) StopPodSandBox(pod *protocol.Pod) error {
	// 停止本Pod的所有容器，包括pause容器；然后回收PodIP等动态资源
	// 因为pause容器的信息没有被记录到Pod里，应该用label来查找
	// 如果后续要重启这个Pod，它在概念上还是原来那一个，但是底层的容器全部发生了改变！
	ctrsLabelKV := GetCtrLabelFilterFromPodConfig(&pod.Config, "")
	// 首先查看这个Pod是否有pause容器，如果没有说明它是一个已移除的，不管它即可；请特别注意此处的写法！
	// 为了少发一次请求，我们先把本Pod的所有容器读出来
	ctrsFilterMap := make(map[string][]string)
	for k, v := range ctrsLabelKV {
		if k != constant.CtrLabel_IsPause { // 添加Pod级别的筛选条件
			ctrsFilterMap["label"] = append(ctrsFilterMap["label"], k+"="+v)
		}
	}
	// 健壮性，应该只停止正在运行的容器
	ctrsFilterMap["status"] = []string{constant.CtrStatus_Running}

	ctrs, err := r.ListContainers(ctrsFilterMap)
	if err != nil {
		logger.KError("Failed to list running containers: %v", err)
		return err
	}
	if len(ctrs) == 0 {
		logger.KInfo("No running container found for podInfo %v, maybe no need to stop", ctrsLabelKV)
		return nil
	}

	// 先停掉其他的所有正在运行的容器
	pauseCtrIds := []string{}
	for _, ctr := range ctrs {
		if ctr.Labels[constant.CtrLabel_IsPause] == constant.CtrLabelVal_IsPauseFalse { // 检测到其他容器，停掉它
			_ = r.StopContainer(ctr.ID)
		} else { // 检测到pause容器，记录下来
			pauseCtrIds = append(pauseCtrIds, ctr.ID)
		}
	}

	if len(pauseCtrIds) == 0 {
		logger.KInfo("No running pause container found for podInfo %v, maybe no need to stop", ctrsLabelKV)
		return nil
	}

	// 这里按道理只有一个pause，但是考虑到某些时候运行失败，希望更彻底地清理掉，所以写成若干个pause的形式
	for _, pauseCtrId := range pauseCtrIds {
		// 停掉pause容器
		_ = r.StopContainer(pauseCtrId)
	}

	// 回写pod状态
	pod.Status = protocol.PodStatus{
		Phase: constant.PodPhaseSucceeded, // 成功终止所有容器
		IP:    "",
	}
	// 如果发生了错误，可能需要在外面额外回写状态而不是在这里
	return nil
}

func (r *RemoteRuntimeService) RemovePodSandBox(pod *protocol.Pod) error {
	// 先尝试停掉这个Pod的所有容器
	err := r.StopPodSandBox(pod)
	if err != nil {
		logger.KError("In RemovePodSandBox, Failed to stop pod: %v", err)
		return err
	}

	// 然后依次删掉它们
	ctrsLabelKV := GetCtrLabelFilterFromPodConfig(&pod.Config, "")
	ctrsFilterMap := make(map[string][]string)
	for k, v := range ctrsLabelKV {
		ctrsFilterMap["label"] = append(ctrsFilterMap["label"], k+"="+v)
	}
	ctrs, err := r.ListContainers(ctrsFilterMap)
	if err != nil {
		logger.KError("Failed to list containers: %v", err)
		return err
	}

	pauseCtrIds := []string{}
	for _, ctr := range ctrs {
		if ctr.Labels[constant.CtrLabel_IsPause] == constant.CtrLabelVal_IsPauseFalse { // 检测到其他容器，删掉它
			_ = r.RemoveContainer(ctr.ID)
		} else { // 检测到pause容器，记录下来
			pauseCtrIds = append(pauseCtrIds, ctr.ID)
		}
	}

	if len(pauseCtrIds) == 0 {
		logger.KInfo("No pause container found for podInfo %v, maybe no need to remove", ctrsLabelKV)
		return nil
	}

	// 移除pause容器
	for _, pauseCtrId := range pauseCtrIds {
		_ = r.RemoveContainer(pauseCtrId)
	}
	// 认为这里不管回写状态
	return nil
}

// 查找本Pod对应的pause容器，返回它的状态；这里CRI的描述很模糊，我们只实现为inspect这个容器
func (r *RemoteRuntimeService) PodSandBoxStatus(pod *protocol.Pod) (*types.ContainerJSON, error) {
	// 通过label找到pause容器
	ctrsLabelKV := GetCtrLabelFilterFromPodConfig(&pod.Config, constant.CtrLabelVal_IsPauseTrue)
	ctrsFilterMap := make(map[string][]string)
	for k, v := range ctrsLabelKV {
		ctrsFilterMap["label"] = append(ctrsFilterMap["label"], k+"="+v)
	}
	ctrs, err := r.ListContainers(ctrsFilterMap)
	if err != nil {
		logger.KError("Failed to list containers: %v", err)
		return nil, err
	}

	// 要求在没有这个pause容器时，返回一个error
	if len(ctrs) == 0 {
		err = errors.New("no pause container found for the pod")
		logger.KError("Failed to find pause container: %v", err)
		return nil, err
	}

	// 认为正常只有一个pause容器，直接返回它的运行状态
	return r.InspectContainer(ctrs[0].ID)
}

// 只在ListContainers之上，加入了得到的结果必须是pause容器的条件；比如可能希望获取某个命名空间下的所有pause容器，那么指定过滤条件为{label:{"namespace=xxx",}}
func (r *RemoteRuntimeService) ListPodSandBox(filter map[string][]string) ([]types.Container, error) {
	// 通过label找到pause容器
	ctrs, err := r.ListContainers(filter)
	if err != nil {
		logger.KError("Failed to list containers: %v", err)
		return nil, err
	}

	// 过滤掉非pause容器
	pauseCtrs := []types.Container{}
	for _, ctr := range ctrs {
		if ctr.Labels[constant.CtrLabel_IsPause] == constant.CtrLabelVal_IsPauseTrue {
			pauseCtrs = append(pauseCtrs, ctr)
		}
	}
	return pauseCtrs, nil
}
