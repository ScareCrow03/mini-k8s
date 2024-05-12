package podtest

import (
	"io/fs"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"mini-k8s/pkg/utils/uid"
	yamlParse "mini-k8s/pkg/utils/yaml"
	"os"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

var test_service *rtm.RemoteRuntimeService

func TestMain(m *testing.M) {
	// 创建一个新的远程运行时服务；这里的超时时间是5分钟，但它并没有被写到任何操作逻辑中，目前仅作为一个摆设
	test_service = rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer test_service.Close()

	m.Run()
	// test_service.ImgSvc.RemoveImageByName(constant.TestContainerImage)
}

func TestCreatePod(t *testing.T) {
	var pod1 protocol.Pod
	//pod1.Config.YAMLToPodConfig("../../../../assets/pod_config_test1.yaml")
	yamlParse.YAMLParse(&pod1.Config, "../../../../assets/pod_config_test1.yaml")
	pod1.Config.Metadata.UID = "mini-k8s_test-uid" + uid.NewUid()
	// 先在当前目录创建一个./test_html.yml文件
	os.Create("/tmp/test_html.html")
	str := "Welcome to nginx"
	os.WriteFile("/tmp/test_html.html", []byte(str), fs.FileMode(os.O_TRUNC))

	err := test_service.CreatePod(&pod1)

	if err != nil {
		t.Fatalf("Failed to create pod: %v", err)
	}

	// 查看本pod状态
	podStatus, err := test_service.GetPodStatusById(pod1.Config.Metadata.UID)
	if err != nil {
		t.Fatalf("Failed to get pod status: %v", err)
	}

	// 因为没有启动容器，应该是pending
	if podStatus.Status.Phase != constant.PodPhasePending {
		t.Fatalf("Failed to find pod in pending status")
	}

	// 启动本pod所有容器
	err = test_service.StartPod(&pod1)
	if err != nil {
		t.Fatalf("Failed to start pod: %v", err)
	}

	// allPodStatus, _ := test_service.GetAllPodsStatusOnNode()
	// data1, _ := yaml.Marshal(&allPodStatus)
	// t.Logf("all Pod status: %s", string(data1))

	// 查看本pod状态
	podStatus, err = test_service.GetPodStatusById(pod1.Config.Metadata.UID)
	if err != nil {
		t.Fatalf("Failed to get pod status: %v", err)
	}

	// 启动完成，应该是running
	if podStatus.Status.Phase != constant.PodPhaseRunning {
		t.Fatalf("Failed to find pod in running status")
	}
	// 打印一下
	data, err := yaml.Marshal(&podStatus)
	if err != nil {
		t.Fatalf("Failed to marshal pod status: %v", err)
	}
	t.Logf("Pod status: %s", string(data))

	// 停止本pod
	err = test_service.StopPodSandBox(&pod1)
	if err != nil {
		t.Fatalf("Failed to stop pod: %v", err)
	}

	// 查看本pod状态
	podStatus, err = test_service.GetPodStatusById(pod1.Config.Metadata.UID)
	if err != nil {
		t.Fatalf("Failed to get pod status: %v", err)
	}

	// 停止完成，要么是failed，要么是succeeded
	if podStatus.Status.Phase != constant.PodPhaseFailed && podStatus.Status.Phase != constant.PodPhaseSucceeded {
		t.Fatalf("Failed to find pod in running status")
	}

	// 移除它
	err = test_service.RemovePodSandBox(&pod1)
	if err != nil {
		t.Fatalf("Failed to remove pod: %v", err)
	}

	// 查看本pod状态，应该不能找到，为一个空体
	podStatus, err = test_service.GetPodStatusById(pod1.Config.Metadata.UID)
	if err != nil {
		t.Logf("Failed to get pod status: %v", err)
	}

	if podStatus.Config.Metadata.UID != "" {
		t.Fatalf("Failed to remove pod")
	}
}
