package containertest

import (
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	weaveClient "mini-k8s/pkg/utils/cni/weave"
	"strings"
	"testing"
	"time"
)

var test_service *rtm.RemoteRuntimeService

func setup() {
	err := test_service.RemoveContainerByNameAndItsImage(constant.TestContainerName, false) // 这个函数如果找不到对应的容器就停下了，所以删容器和镜像的操作建议分开
	if err != nil {
		fmt.Printf("Failed to remove container and its image: %v\n, maybe no need to remove in setup", err)
	}
}

func TestMain(m *testing.M) {
	// 创建一个新的远程运行时服务；这里的超时时间是5分钟，但它并没有被写到任何操作逻辑中，目前仅作为一个摆设
	test_service = rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer test_service.Close()

	m.Run()
	test_service.ImgSvc.RemoveImageByName(constant.TestContainerImage)
}

func TestWeaveAttach(t *testing.T) {
	setup()
	// 创建一个新的容器，指定一些配置
	containerConfig := &protocol.ContainerConfig{
		// TODO: 待指定更多配置
		Name:            constant.TestContainerName,
		Image:           constant.TestContainerImage,
		Command:         []string{"sh", "-c", "while true; do sleep 1; done"}, // 采用alpine镜像时，必须启动一个长期循环，防止container在启动后立即退出（这是alpine的默认行为）
		ImagePullPolicy: protocol.PullIfNotPresentStr,
	}

	containerID, err := test_service.CreateContainer(containerConfig)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// 启动容器
	err = test_service.StartContainer(containerID)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	// 以下测试weave功能

	// 为它分配一个IP
	ip1, err := weaveClient.AttachCtr(containerID)
	if err != nil {
		t.Fatalf("Failed to attach container to weave network: %v", err)
	}
	fmt.Printf("container get weave IP: %s\n", ip1)

	// Lookup
	ip2, err := weaveClient.LookupIP(containerID)
	if err != nil {
		t.Fatalf("Failed to lookup IP by container ID: %v", err)
	}

	if strings.Compare(ip1, ip2) != 0 {
		t.Fatalf("IPs are not equal: %s, %s", ip1, ip2)
	}

	// detach
	err = weaveClient.DetachCtr(containerID)
	if err != nil {
		t.Fatalf("Failed to detach container from weave network: %v", err)
	}
	fmt.Printf("container detached from weave network\n")

	// Lookup
	ip3, err := weaveClient.LookupIP(containerID)
	if err != nil {
		t.Fatalf("Failed to lookup IP by container ID: %v", err)
	}
	if ip3 != "" {
		t.Fatalf("IP is not empty: %s", ip3)
	}
}

func TestCreateContainer(t *testing.T) {
	setup()
	// 创建一个新的容器，指定一些配置
	containerConfig := &protocol.ContainerConfig{
		// TODO: 待指定更多配置
		Name:            constant.TestContainerName,
		Image:           constant.TestContainerImage,
		Command:         []string{"sh", "-c", "while true; do sleep 1; done"}, // 采用alpine镜像时，必须启动一个长期循环，防止container在启动后立即退出（这是alpine的默认行为）
		ImagePullPolicy: protocol.AlwaysPullStr,
		Labels:          map[string]string{"test": "test"},
	}

	containerID, err := test_service.CreateContainer(containerConfig)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// 启动容器
	err = test_service.StartContainer(containerID)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// 获取容器列表，按name过滤，查看当前测试容器是否存在
	containers, err := test_service.ListContainers(map[string][]string{"name": {constant.TestContainerName},
		"label": {"test=test"}}) // 请注意按照label过滤的写法
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}
	fmt.Printf("container list: %v\n", containers)
	if len(containers) == 0 {
		t.Fatalf("Unexpected container list: %v", containers)
	}

	// 获取容器的信息
	_, err = test_service.InspectContainer(containerID)
	if err != nil {
		t.Fatalf("Failed to inspect container: %v", err)
	}

	// 获取容器的状态
	_, err = test_service.ContainerStatus(containerID)
	if err != nil {
		t.Fatalf("Failed to get container status: %v", err)
	}

	// 在容器中执行一段指令
	output, err := test_service.ExecContainer(containerID, []string{"echo", "Hello, World!"})
	if err != nil {
		t.Fatalf("Failed to execute command in container: %v", err)
	}
	output = strings.TrimSpace(output) // 删除输出的开头和结尾的空白字符，但还是有可能有奇怪的东西，那么我们只能检测它是否包含我们期望的字符串
	if !strings.Contains(output, "Hello, World!") {
		t.Fatalf("Unexpected output: " + output)
	}

	// 停止容器
	err = test_service.StopContainer(containerID)
	if err != nil {
		t.Fatalf("Failed to stop container: %v", err)
	}

	time.Sleep(3 * time.Second) // 等待一会
	// 移除容器
	err = test_service.RemoveContainer(containerID)
	if err != nil {
		t.Fatalf("Failed to remove container: %v", err)
	}
}
