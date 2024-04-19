package remoteRuntime

import (
	"mini-k8s/pkg/protocol"
	"strings"
	"testing"
	"time"
)

var (
	testContainerName  = "mini-k8s-test-alpine"
	testContainerImage = "alpine:latest" // 这个测试镜像应该现在没有人正在用！否则可能有其他问题
)

var service *remoteRuntimeService

func TestMain(m *testing.M) {
	// 创建一个新的远程运行时服务
	service = NewRemoteRuntimeService(5 * time.Minute)
	defer service.Close()
	// 建议先把镜像删掉！
	// service.RemoveImageByNameAndItsContainers(testContainerImage)
	service.RemoveContainerByNameAndItsImage(testContainerName, false) // 这个函数如果找不到对应的容器就停下了，所以删容器和镜像的操作建议分开
	service.ImgSvc.RemoveImageByName(testContainerImage)
	m.Run()
	service.RemoveContainerByNameAndItsImage(testContainerName, true)
	service.ImgSvc.RemoveImageByName(testContainerImage)
}

func TestRemoteRuntimeService(t *testing.T) {
	// 创建一个新的容器，指定一些配置
	containerConfig := &protocol.ContainerConfig{
		// TODO: 待指定更多配置
		Name:  testContainerName,
		Image: testContainerImage,
		Cmd:   []string{"sh", "-c", "while true; do sleep 1; done"}, // 采用alpine镜像时，必须启动一个长期循环，防止container在启动后立即退出（这是alpine的默认行为）
	}

	containerID, err := service.CreateContainer(containerConfig, protocol.AlwaysPull)
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
	output = strings.TrimSpace(output) // 删除输出的开头和结尾的空白字符，但还是有可能有奇怪的东西，那么我们只能检测它是否包含我们期望的字符串
	if !strings.Contains(output, "Hello, World!") {
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
