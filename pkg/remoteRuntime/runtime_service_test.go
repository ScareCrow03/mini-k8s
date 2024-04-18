package remoteRuntime

import (
	"mini-k8s/pkg/protocol"
	"testing"
	"time"
)

var (
	testContainerName  = "mini-k8s-test-redis"
	testContainerImage = "docker.io/library/redis:latest"
)

var service *remoteRuntimeService

func TestMain(m *testing.M) {
	// 创建一个新的远程运行时服务
	service = NewRemoteRuntimeService(5 * time.Minute)
	defer service.Close()
	// 请先把镜像删掉；此时应该没有别人正在使用这个镜像！
	service.RemoveImageAndContainers(testContainerImage)
	m.Run()
	service.RemoveImageAndContainers(testContainerImage)
}

func TestRemoteRuntimeService(t *testing.T) {

	// 创建一个新的容器，指定一些配置
	containerConfig := &protocol.ContainerConfig{
		// TODO: 填写待测容器配置
		Name:  testContainerName,
		Image: testContainerImage,
	}
	containerID, err := service.CreateContainer(containerConfig, true)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// 启动容器
	err = service.StartContainer(containerID)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// 获取容器列表，按name过滤，查看当前测试容器是否存在
	containers, err := service.ListContainers(map[string][]string{"name": {testContainerName}})
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("Unexpected container list: %v", containers)
	}

	// 获取容器的信息
	_, err = service.InspectContainer(containerID)
	if err != nil {
		t.Fatalf("Failed to inspect container: %v", err)
	}

	// 获取容器的状态
	_, err = service.ContainerStatus(containerID)
	if err != nil {
		t.Fatalf("Failed to get container status: %v", err)
	}

	// 在容器中执行一段指令
	output, err := service.ExecContainer(containerID, []string{"echo", "Hello, World!"})
	if err != nil {
		t.Fatalf("Failed to execute command in container: %v", err)
	}
	if output != "Hello, World!" {
		t.Fatalf("Unexpected output: " + output)
	}

	// 停止容器
	err = service.StopContainer(containerID)
	if err != nil {
		t.Fatalf("Failed to stop container: %v", err)
	}

	// 移除容器
	err = service.RemoveContainer(containerID)
	if err != nil {
		t.Fatalf("Failed to remove container: %v", err)
	}
}
