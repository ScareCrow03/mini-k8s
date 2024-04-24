package remoteRuntime

// func (r *remoteRuntimeService) CreatePauseContainer(cfg *protocol.PodConfig) (string, error) {
// 	// 创建一个特殊的 ContainerConfig 用于 pause 容器
// 	pauseContainerConfig := protocol.ContainerConfig{
// 		Name:  "pause",
// 		Image: "registry.aliyuncs.com/google_containers/pause:3.9", // pause 容器的镜像
// 		// 其他必要的配置
// 	}

// 	// // 将 pause 容器的配置添加到 PodConfig 中
// 	// cfg.Spec.Containers = append([]protocol.ContainerConfig{pauseContainerConfig}, cfg.Spec.Containers...)

// 	// 调用 CreateContainer 方法创建 pause 容器
// 	pauseContainerID, err := r.CreateContainer(&pauseContainerConfig)
// 	if err != nil {
// 		return "", err
// 	}

// 	// // 将 pause 容器的 ID 保存到 PodConfig 的 Metadata 中
// 	// cfg.Metadata.UID = pauseContainerID

// 	return pauseContainerID, nil
// }

// func TestCreatePause(t *testing.T) {
// 	// 创建一个 PodConfig 对象
// 	podConfig := protocol.PodConfig{
// 		// PodConfig 的其他配置
// 	}

// 	// 调用 CreatePauseContainer 方法创建 pause 容器
// 	pauseContainerID, err := test_service.CreatePauseContainer(&podConfig)
// 	if err != nil {
// 		t.Fatalf("Failed to create pause container: %v", err)
// 	}
// 	test_service.StartContainer(pauseContainerID)

// 	// 建立两个新的容器，加入它
// 	containerConfig1 := &protocol.ContainerConfig{
// 		// TODO: 待指定更多配置
// 		Name:               testContainerName,
// 		Image:              "nginx:1.19.0",
// 		ImagePullPolicy:    "IfNotPresent",
// 		NetworkContainerID: pauseContainerID,
// 	}
// 	containerConfig1.Ports = []struct {
// 		Name          string `yaml:"name"`
// 		ContainerPort int64  `yaml:"containerPort"`
// 		HostPort      int64  `yaml:"hostPort"`
// 		Protocol      string `yaml:"protocol"`
// 	}{
// 		{
// 			Name:          "http",
// 			ContainerPort: 90,
// 			HostPort:      11223,
// 			Protocol:      "tcp",
// 		},
// 	}

// 	containerConfig2 := &protocol.ContainerConfig{
// 		// TODO: 待指定更多配置
// 		Name:               testContainerName + "2",
// 		Image:              "curlimages/curl:latest",
// 		Command:            []string{"sh", "-c", "while true; do sleep 1; done"}, // 采用alpine镜像时，必须启动一个长期循环，防止container在启动后立即退出（这是alpine的默认行为）
// 		ImagePullPolicy:    "IfNotPresent",
// 		NetworkContainerID: pauseContainerID,
// 	}

// 	// 创建容器
// 	id1, _ := test_service.CreateContainer(containerConfig1)

// 	id2, _ := test_service.CreateContainer(containerConfig2)

// 	// 启动容器
// 	test_service.StartContainer(id1)
// 	test_service.StartContainer(id2)

// 	// 查看id2能否用curl访问id1，使用的是localhost:80
// 	outStr, _ := test_service.ExecContainer(id2, []string{"curl", "localhost:90"})
// 	fmt.Printf("container 2 curl localhost:80: %s\n", outStr)
// }
