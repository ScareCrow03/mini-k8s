package rtm

import (
	"errors"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	flannelClient "mini-k8s/pkg/utils/cni/flannel"

	"time"

	"github.com/docker/docker/api/types"
)

// 这个文件用于实现Pod级别的方法，包括按照Pod配置的数据结构，把该Pod真实运行起来，以及对Pod的相关操作；它依赖于更低层的sandbox, container, image等服务
func (r *RemoteRuntimeService) CreatePod(pod *protocol.Pod) error {
	// 创建pause容器，作为整个Pod的沙盒
	pauseId, err := r.RunPodSandBox(pod)
	if err != nil {
		logger.KError("Failed to create pod sandbox: %v", err)
	}

	// 创建Pod中的其他容器；首先需要考虑挂载卷映射的问题
	// 方法是，先提取出pod配置记录的name-hostPath映射关系
	// 遍历spec容器配置，将containerPath根据volume名字，映射到hostPath
	// 我们默认用户指定的主机路径都是合法的
	volumeName2HostPath := make(map[string]string)
	for _, volume := range pod.Config.Spec.Volumes {
		volumeName2HostPath[volume.Name] = volume.HostPath.Path
	}

	ctrIds := make([]string, 0)

	for _, containerConfig := range pod.Config.Spec.Containers {
		// 挂载卷映射
		for _, volumeMount := range containerConfig.VolumeMounts {
			if volumeName2HostPath[volumeMount.Name] == "" {
				// 指定了不存在的共享卷名，在底层docker SDK创建容器时会跳过它！这里只是提出一个警告
				logger.KWarning("For container %s, Volume name %s not found in pod %v, skip it", containerConfig.Name, volumeMount.Name, pod.Config.Metadata)
			}
		}

		// 在Pod中创建容器，需要指定pauseId，以及volumeName到hostPath的映射关系，以及Pod配置自身（用于设置Label）
		ctrId, err := r.CreateContainerInPod(&containerConfig, pauseId, &volumeName2HostPath, pod, constant.CtrLabelVal_IsPauseFalse)
		if err != nil {
			logger.KError("Failed to create container: %v", err)
			return err
		}
		ctrIds = append(ctrIds, ctrId)
	}

	// 回写Container ID，注意需要是引用！
	ctrsCfg := &pod.Config.Spec.Containers
	for i := range *ctrsCfg {
		(*ctrsCfg)[i].UID = ctrIds[i]
	}
	fmt.Printf("Now Pod ip: %s\n", pod.Status.IP)
	return nil
}

func (r *RemoteRuntimeService) StartPod(pod *protocol.Pod) error {
	// 只有在Create之后才能Start
	_, otherCtrs, err := r.ListPodContainersById(pod.Config.Metadata.UID)
	if err != nil {
		logger.KError("Failed to list pod containers: %v", err)
		return err
	}

	// 因为pause在正常情况下只在Create时建立并启动，所以不管它
	for _, ctr := range otherCtrs {
		err = r.StartContainer(ctr.ID)
		if err != nil {
			logger.KError("Failed to start container: %v", err)
			return err
		}
	}
	return nil
}

// 传递pod对象的Stop方法直接在Sandbox层面实现好，调用一遍即可
// 传递pod对象的Remove方法直接在Sandbox层面实现好，调用一遍即可

// 列出Pod中的所有容器，返回值1是pause容器，返回值2是其他，返回值3是错误信息；为了健壮性，这里pause仍然开的是数组，但正常情况下应该只有1个！
func (r *RemoteRuntimeService) ListPodContainersById(podId string) ([]types.Container, []types.Container, error) {
	// 只需要添加一个label:podId的过滤条件
	ctrsFilterMap := make(map[string][]string)
	ctrsFilterMap["label"] = append(ctrsFilterMap["label"], constant.CtrLabel_PodId+"="+podId)

	ctrs, err := r.ListContainers(ctrsFilterMap)
	if err != nil {
		logger.KError("Failed to list containers: %v", err)
		return nil, nil, err
	}

	pauseCtrs := make([]types.Container, 0)
	otherCtrs := make([]types.Container, 0)
	for _, ctr := range ctrs {
		if ctr.Labels[constant.CtrLabel_IsPause] == constant.CtrLabelVal_IsPauseTrue {
			pauseCtrs = append(pauseCtrs, ctr)
		} else {
			otherCtrs = append(otherCtrs, ctr)
		}
	}

	return pauseCtrs, otherCtrs, nil
}

func (r *RemoteRuntimeService) GetPodStatusById(podId string) (*protocol.Pod, error) {
	pod := &protocol.Pod{}

	pauseCtrs, otherCtrs, err := r.ListPodContainersById(podId)
	if err != nil {
		logger.KError("Failed to list pod containers: %v", err)
		return nil, err
	}

	if len(pauseCtrs) == 0 { // 返回一个空的pod状态结构体
		logger.KInfo("No pause container found for pod %s", podId)
		return pod, nil
	} else if len(pauseCtrs) > 1 {
		err = errors.New("more than one pause container found for the pod")
		return pod, err
	}

	// 有且只有一个pause容器
	pauseCtr := pauseCtrs[0]
	pod.Config.Metadata = protocol.MetadataType{
		Name:      pauseCtr.Labels[constant.CtrLabel_PodName],
		Namespace: pauseCtr.Labels[constant.CtrLabel_PodNamespace],
		UID:       podId,
	}

	podIP, err := flannelClient.LookupIP(pauseCtr.ID)
	if err != nil {
		logger.KError("failed to lookup IP by container ID: %v", err)
		return nil, err
	}
	// 填写与pause容器相关的信息
	pod.Status = protocol.PodStatus{
		IP:         podIP,
		Runtime:    time.Since(time.Unix(pauseCtr.Created, 0)),
		UpdateTime: time.Now(),
	}

	// map必须先做初始化
	if pod.Status.ContainerStatus == nil {
		pod.Status.ContainerStatus = make(map[string]types.ContainerState)
	}

	if pod.Status.CtrsMetrics == nil {
		pod.Status.CtrsMetrics = make(map[string]protocol.CtrMetricsEntry)
	}

	// 收集每个容器的状态，使用inspect
	ctrsStatus := make([]types.ContainerState, 0)
	for _, ctr := range otherCtrs {
		ctrStatusJSON, err := r.InspectContainer(ctr.ID)
		if err != nil {
			logger.KError("Failed to inspect container: %v", err)
			continue
		}

		pod.Status.ContainerStatus[ctr.ID] = *ctrStatusJSON.State

		ctrStatsJson, err := r.ContainerStatus(ctr.ID)
		if err != nil {
			logger.KError("Failed to get container status: %v", err)
			continue
		}
		oneCtrStats := protocol.ParseDockerCtrStatsToMetricsEntry(ctrStatsJson)
		pod.Status.CtrsMetrics[ctr.ID] = oneCtrStats
		ctrsStatus = append(ctrsStatus, *ctrStatusJSON.State)
	}

	// 再计算本pod应该处于什么状态
	pod.Status.Phase = checkPodPhaseFromCtrsStatus(ctrsStatus)
	// 此时所有Pod状态字段已经填好了！

	return pod, nil
}

// 本Node上的所有Pod运行状态查看；这个数据会定期发回给API-SERVER并做更新，用户看到的某个单独的Pod的状态、都是从中提取的
func (r *RemoteRuntimeService) GetAllPodsStatusOnNode() (map[string]*protocol.Pod, error) {
	// 我们直接返回一个pod体，但是这里的pod体是不完整的，只有metadata和status，没有spec
	pods := make(map[string]*protocol.Pod)

	// 首先找到所有pause容器，这里不添加过滤条件指的是本节点上的所有pause容器
	pauseCtrs, err := r.ListPodSandBox(nil)
	if err != nil {
		logger.KError("Failed to list pause containers: %v", err)
		return nil, err
	}

	// 获取podId，添加到map中；这里也可以从pause容器直接知道podIP的信息，从weave获取即可
	for _, ctr := range pauseCtrs {
		podId := ctr.Labels[constant.CtrLabel_PodId]
		if pods[podId] != nil {
			logger.KError("more than one pause container found for the pod %s", podId)
			pods[podId].Status.Phase = constant.PodPhaseUnknown
			continue
		}

		podIP, err := flannelClient.LookupIP(ctr.ID)
		if err != nil {
			logger.KError("failed to lookup IP by container ID: %v", err)
			continue
		}

		pods[podId] = new(protocol.Pod) // 初始化pod指针，否则下一行会为nil赋值

		pods[podId].Config.Metadata = protocol.MetadataType{
			Name:      ctr.Labels[constant.CtrLabel_PodName],
			Namespace: ctr.Labels[constant.CtrLabel_PodNamespace],
			UID:       podId,
		}
		// pod的运行时间根据pause容器的创建时间来计算；注意Created字段是64位时间戳，从1970年1月1日到现在的秒数；更新时间为现在
		pods[podId].Status = protocol.PodStatus{
			IP:         podIP,
			Runtime:    time.Since(time.Unix(ctr.Created, 0)),
			UpdateTime: time.Now(),
		}
	}

	// 然后按每个podId，找到对应的其他容器，计算相应状态
	for podId := range pods {
		if pods[podId].Status.Phase != constant.PodPhaseUnknown {
			continue
		}

		_, otherCtrs, err := r.ListPodContainersById(podId)
		if err != nil {
			logger.KError("Failed to list pod containers: %v", err)
			return nil, err
		}

		// map对象必须先初始化才能用
		pods[podId].Status.ContainerStatus = make(map[string]types.ContainerState)
		pods[podId].Status.CtrsMetrics = make(map[string]protocol.CtrMetricsEntry)

		// 收集每个容器的配置状态，使用inspect
		simpleCtrsState := make([]types.ContainerState, 0)
		ctrsStats := make([]protocol.CtrMetricsEntry, 0)
		for _, ctr := range otherCtrs {
			ctrStateJSON, err := r.InspectContainer(ctr.ID)
			if err != nil {
				logger.KError("Failed to inspect container: %v", err)
				continue
			}
			// 写回每个容器的running信息
			pods[podId].Status.ContainerStatus[ctr.ID] = *ctrStateJSON.State
			simpleCtrsState = append(simpleCtrsState, *ctrStateJSON.State)

			// 收集每个容器的运行状态
			ctrStatsJson, err := r.ContainerStatus(ctr.ID)
			if err != nil {
				logger.KError("Failed to get container status: %v", err)
				continue
			}
			oneCtrStats := protocol.ParseDockerCtrStatsToMetricsEntry(ctrStatsJson)
			pods[podId].Status.CtrsMetrics[ctr.ID] = oneCtrStats

			ctrsStats = append(ctrsStats, oneCtrStats)
		}

		// 再计算本pod应该处于什么状态
		pods[podId].Status.Phase = checkPodPhaseFromCtrsStatus(simpleCtrsState)
		// 此时所有Pod状态字段已经填好了！

		// 写回每个容器的运行信息，并计算整个Pod的运行信息
		pods[podId].Status.PodMetrics = protocol.CalculatePodMertrics(podId, ctrsStats)
	}
	return pods, nil
}

func checkPodPhaseFromCtrsStatus(ctrsStatus []types.ContainerState) string {
	// 由于现在并不知道本pod的期望config是什么，无法计算出Spec有多少容器
	// 故只能简单认为，当前如果没有容器运行，则为pending（因为我们默认是同步创建每个容器）
	if len(ctrsStatus) == 0 {
		return constant.PodPhasePending
	}

	// 如果所有容器都是created状态，那么就是pending
	allCreated := true
	for _, ctrStatus := range ctrsStatus {
		if ctrStatus.Status != constant.CtrStatus_Created {
			allCreated = false
		}
	}
	if allCreated {
		return constant.PodPhasePending
	}

	// 至少一个容器在运行，那么就是running
	for _, ctrStatus := range ctrsStatus {
		if ctrStatus.Status == constant.CtrStatus_Running {
			return constant.PodPhaseRunning
		}
	}

	// 所有容器终止运行，且有一个容器非正常退出，那么就是failed
	for _, ctrStatus := range ctrsStatus {
		if ctrStatus.Status == constant.CtrStatus_Exited && ctrStatus.ExitCode == 0 {
			return constant.PodPhaseFailed
		}
	}
	// 全部正常退出，那么就是succeeded
	return constant.PodPhaseSucceeded
}
