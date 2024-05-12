package message

var (
	// MsgQueueName is the name of the message queue
	UpdatePodQueueName    = "updatePod"
	CreatePodQueueName    = "createPod"
	KubeletCreatePodQueue = "kubeletCreatePodQueue"
	KubeletStopPodQueue   = "kubeletStopPodQueue"
	KubeletDeletePodQueue = "kubeletDeletePodQueue"
	CreateDnsQueueName    = "createDns"
	UpdateDnsQueueName    = "updateDns"  //apiserver向controller发送dns更新请求
	DeleteDnsQueueName    = "deleteDns"  //apiserver向controller发送dns删除请求
	UpdateHostQueueName   = "updateHost" //apiserver向kubeproxy发送host和conf更新请求

	CreateServiceQueueName = "createServiceQueue"
	DeleteServiceQueueName = "deleteServiceQueue"

	CreateReplicasetQueueName = "createReplicasetQueue"
	DeleteReplicasetQueueName = "deleteReplicasetQueue"
)
