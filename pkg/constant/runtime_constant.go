package constant

const (
	CtrPauseImgUrl      = "registry.aliyuncs.com/google_containers/pause:3.9"
	CtrPauseName_Prefix = "pause-" // pause容器的名字前缀

	CtrLabel_PodId        = "pod_id"
	CtrLabel_PodNamespace = "pod_ns"
	CtrLabel_PodName      = "pod_name"
	CtrLabel_IsPause      = "is_pause"

	CtrLabelVal_IsPauseTrue  = "true"
	CtrLabelVal_IsPauseFalse = "false"

	TestContainerImage = "alpine:latest"
	TestContainerName  = "mini-k8s-test-alpine"

	CtrStatus_Created = "created"
	CtrStatus_Running = "running"
	CtrStatus_Exited  = "exited"
	CtrStatus_Stopped = CtrStatus_Exited // 注意docker里描述停止的容器为退出

	PodPhasePending   = "Pending"
	PodPhaseRunning   = "Running"
	PodPhaseSucceeded = "Succeeded"
	PodPhaseFailed    = "Failed"
	PodPhaseUnknown   = "Unknown"
	// yaml中大小写任意
	PodRestartPolicyAlways    = "always"
	PodRestartPolicyNever     = "never"
	PodRestartPolicyOnFailure = "onfailure"
)
