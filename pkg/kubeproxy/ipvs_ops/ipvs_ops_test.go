package ipvs_ops

import (
	"fmt"
	"io/fs"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/protocol/service_cfg"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	yamlParse "mini-k8s/pkg/utils/yaml"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

const IPVS_TEST_UID_PREFIX = "mini-k8s_test_ipvs-uid"

func TestMain(m *testing.M) {
	m.Run()
}

func TestClearAll(t *testing.T) {
	ops := NewIpvsOps(constant.CLUSTER_CIDR_DEFAULT)
	defer ops.Close()
	ops.Clear()

	rtm_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer rtm_service.Close()

	var pod1 protocol.Pod
	yamlParse.YAMLParse(&pod1.Config, "../../../assets/test_ipvs/test_ipvs_pod1.yaml")
	pod1.Config.Metadata.UID = IPVS_TEST_UID_PREFIX + "1"
	rtm_service.RemovePodSandBox(&pod1)
	// hello
}

func TestClusterIP(t *testing.T) {
	// 创建一个./test_ipvs.html文件，写一点东西进去，后面被nginx用到
	os.Create("./test_ipvs.html")
	str := "Welcome to nginx"
	os.WriteFile("./test_ipvs.html", []byte(str), fs.FileMode(os.O_TRUNC))

	// 读取yaml文件，获取Pod和Service的配置信息
	var pod1 protocol.Pod
	yamlParse.YAMLParse(&pod1.Config, "../../../assets/test_ipvs/test_ipvs_pod1.yaml")
	pod1.Config.Metadata.UID = IPVS_TEST_UID_PREFIX + "1"

	var svc1 service_cfg.ServiceType
	yamlParse.YAMLParse(&svc1.Config, "../../../assets/test_ipvs/test_ipvs_service1.yaml")
	svc1.Config.Metadata.UID = IPVS_TEST_UID_PREFIX + "2"

	// 1. 创建并运行一个pod
	rtm_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer rtm_service.Close()
	rtm_service.CreatePod(&pod1)
	rtm_service.StartPod(&pod1)

	// 查看各endPoint信息；这个函数仅用于查看PodIP，其他信息都需要来自于原始的spec！
	pod1_status, _ := rtm_service.GetPodStatusById(pod1.Config.Metadata.UID)
	data, _ := yaml.Marshal(&pod1_status)
	fmt.Printf("Pod1 status: %s", string(data))
	// 配置service的endpoints信息
	// 这个函数可以单独提取出来服用
	for _, container := range pod1.Config.Spec.Containers {
		// 如果这个容器暴露的端口不为空
		if len(container.Ports) > 0 {
			// 逐一绑定
			for _, port := range container.Ports {
				ep := service_cfg.Endpoint{
					PodUID: pod1.Config.Metadata.UID,
					IP:     pod1_status.Status.IP,
					Port:   int(port.ContainerPort),
				}
				svc1.Status.Endpoints = append(svc1.Status.Endpoints, ep)
			}
		}

	}

	// 2. 创建一个service
	ops := NewIpvsOps(constant.CLUSTER_CIDR_DEFAULT)
	ops.Init()
	defer ops.Close()

	svc1.Config.Spec.ClusterIP = "222.111.0.0"

	data, _ = yaml.Marshal(&svc1)
	fmt.Printf("Svc status: %s", string(data))
	ops.AddService(&svc1)
	ops.SaveToFile(constant.IPTABLES_FILE_PATH, constant.IPVS_FILE_PATH, constant.IPSET_FILE_PATH)

	// 3. 模拟向这个Service的80端口发消息，因为是nginx的默认回复应该有Welcome信息
	output, err := exec.Command("curl", "--max-time", "3", svc1.Config.Spec.ClusterIP+":80").Output()
	if !strings.Contains(string(output), "Welcome") {
		t.Fatalf("Error curl Service addr"+svc1.Config.Spec.ClusterIP+":80, reason: %s", err)
	}
}
