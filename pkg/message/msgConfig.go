package message

var (
	// MsgQueueName is the name of the message queue
	UpdatePodQueueName    = "updatePod"
	CreatePodQueueName    = "createPod"
	KubeletCreatePodQueue = "kubeletCreatePodQueue"
	KubeletStopPodQueue   = "kubeletStopPodQueue"
	KubeletDeletePodQueue = "kubeletDeletePodQueue"
)
