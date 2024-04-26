package testpod

import (
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"mini-k8s/pkg/utils/yaml"
	"testing"
	"time"
)

var test_service *rtm.RemoteRuntimeService

func setup() {
	test_service.RemoveContainerByNameAndItsImage(constant.TestContainerName, false) // 这个函数如果找不到对应的容器就停下了，所以删容器和镜像的操作建议分开
}

func TestMain(m *testing.M) {
	// 创建一个新的远程运行时服务；这里的超时时间是5分钟，但它并没有被写到任何操作逻辑中，目前仅作为一个摆设
	test_service = rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer test_service.Close()

	m.Run()
}

func TestCreateSandbox(t *testing.T) {
	setup()

	var pod1 protocol.Pod
	//pod1.Config.YAMLToPodConfig("../../../../assets/pod_config_test1.yaml")
	yamlParse.YAMLParse(&pod1.Config, "../../../../assets/pod_config_test1.yaml")

	// 认为uid是在外面产生的
	pod1.Config.Metadata.UID = "mini-k8s_test-uid" + uid.NewUid()
	pauseId1, err := test_service.RunPodSandBox(&pod1)
	if err != nil {
		t.Fatalf("Failed to create sandbox: %v", err)
	}

	pauseStatus, err := test_service.PodSandBoxStatus(&pod1)
	if err != nil {
		t.Fatalf("Failed to get sandbox status: %v", err)
	}

	// 查看是否找到了本pause容器
	if pauseStatus.ID != pauseId1 {
		t.Fatalf("Failed to find pause container")
	}

	// 重新run一遍，它应该要删掉旧Pod的所有容器，然后创建新的
	pod1_new := pod1
	pauseId1_new, err := test_service.RunPodSandBox(&pod1_new) // 注意pod1和pod2并不是同一个结构体；但是它们的config指定相同，则run的时候视为同一个概念上的pod，会先删除掉旧状态Pod的所有容器、然后创建一份新的；而且赋值新的status的目标是pod1_new，现在pod1保存的都是陈旧状态的数据
	if err != nil {
		t.Fatalf("Failed to create sandbox: %v", err)
	}
	// 因为这里在测试的时候，只启动了一个pod，那么IP在重启后可能是一样的，这是无所谓的，我们在底层确实将它释放了！

	// 重新run之后，此时这个pod的pause容器应该是新的
	if pauseId1 == pauseId1_new {
		t.Fatalf("Failed to create new sandbox due to pause container conflict")
	}

	// 移除新的pod
	err = test_service.RemovePodSandBox(&pod1_new)
	if err != nil {
		t.Fatalf("Failed to stop sandbox: %v", err)
	}
}
