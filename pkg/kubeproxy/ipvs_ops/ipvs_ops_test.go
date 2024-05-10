package ipvs_ops

import (
	"fmt"
	"io/fs"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	"mini-k8s/pkg/utils/net_util"
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
	ClearAllPodsAndRules()
	m.Run()
}

func ClearAllPodsAndRules() {
	ops := NewIpvsOps(constant.CLUSTER_CIDR_DEFAULT)
	defer ops.Close()
	ops.Clear()

	rtm_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer rtm_service.Close()

	var pod1 protocol.Pod
	yamlParse.YAMLParse(&pod1.Config, "../../../assets/test_ipvs/test_ipvs_pod1.yaml")
	pod1.Config.Metadata.UID = IPVS_TEST_UID_PREFIX + "1"
	rtm_service.RemovePodSandBox(&pod1)

	var pod2 protocol.Pod
	yamlParse.YAMLParse(&pod2.Config, "../../../assets/test_ipvs/test_ipvs_curl_pod.yaml")
	pod2.Config.Metadata.UID = IPVS_TEST_UID_PREFIX + "3"
	rtm_service.RemovePodSandBox(&pod2)
}

func TestClearAll(t *testing.T) {
	ClearAllPodsAndRules()
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

	var svc1 protocol.ServiceType
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
				ep := protocol.Endpoint{
					PodUID: pod1.Config.Metadata.UID,
					IP:     pod1_status.Status.IP,
					Port:   int(port.ContainerPort),
				}
				svc1.Status.Endpoints = append(svc1.Status.Endpoints, ep)
			}
		}
	}

	// 2. 创建一个service，手动指定ClusterIP
	ops := NewIpvsOps(constant.CLUSTER_CIDR_DEFAULT)
	ops.Init()
	defer ops.Close()
	data, _ = yaml.Marshal(&svc1)
	fmt.Printf("Svc status: %s", string(data))
	svc1.Config.Spec.ClusterIP = "222.111.0.0"

	ops.AddService(&svc1)
	ops.SaveToFile(constant.IPTABLES_FILE_PATH, constant.IPVS_FILE_PATH, constant.IPSET_FILE_PATH)

	// 3. 模拟本机向这个Service的80端口发消息，因为是nginx的默认回复应该有Welcome信息
	svc1_clusterIP_addr_str := svc1.Config.Spec.ClusterIP + ":80"
	output, err := exec.Command("curl", "--max-time", "3", svc1_clusterIP_addr_str).Output()
	if !strings.Contains(string(output), "Welcome") {
		t.Fatalf("Error curl Service addr"+svc1.Config.Spec.ClusterIP+":80 from Host Machine, reason: %s", err)
	}

	// 4. 创建一个不同于上述的，包含curl容器的Pod2，向这个SERVICE发消息
	var pod2 protocol.Pod
	yamlParse.YAMLParse(&pod2.Config, "../../../assets/test_ipvs/test_ipvs_curl_pod.yaml")
	pod2.Config.Metadata.UID = IPVS_TEST_UID_PREFIX + "3"
	rtm_service.CreatePod(&pod2)
	rtm_service.StartPod(&pod2)

	var pod2_curl_id string
	for _, container := range pod2.Config.Spec.Containers {
		// 先找到本pod的curl容器
		if strings.Contains(container.Image, "curl") {
			pod2_curl_id = container.UID
		}
	}

	output_str, err := rtm_service.ExecContainer(pod2_curl_id, []string{"curl", "--max-time", "3", svc1_clusterIP_addr_str})
	// 查看是否成功
	if !strings.Contains(output_str, "Welcome") {
		t.Fatalf("Error curl ClusterIP Service addr "+svc1.Config.Spec.ClusterIP+":80 from the same pod (hairpair), reason: %s", err)
	}

	// 5. 模拟本Pod内的curl容器、向该Service自环发消息，这是最强的条件
	var hairpair_curl_id string
	for _, container := range pod1.Config.Spec.Containers {
		// 先找到本pod的curl容器
		if strings.Contains(container.Image, "curl") {
			hairpair_curl_id = container.UID
		}
	}
	output_str, err = rtm_service.ExecContainer(hairpair_curl_id, []string{"curl", "--max-time", "3", svc1_clusterIP_addr_str})
	// 查看是否成功
	if !strings.Contains(output_str, "Welcome") {
		t.Fatalf("Error curl ClusterIP Service addr "+svc1.Config.Spec.ClusterIP+":80 from the same pod (hairpair), reason: %s", err)
	}

	// 6. 删除这个服务，后续应该不能访问到它！
	ops.DelService(&svc1)
	output, err = exec.Command("curl", "--max-time", "3", svc1_clusterIP_addr_str).Output()
	// 应该不成功！
	if string(output) != "" {
		t.Fatalf("Delete Cluster Service addr "+svc1.Config.Spec.ClusterIP+":80, but still can curl it, reason: %s", err)
	}

	// 测试完成，回收资源
	// rtm_service.RemovePodSandBox(&pod1)
	// ops.Clear()
	ClearAllPodsAndRules()
}

func TestNodePort(t *testing.T) {
	// 创建一个./test_ipvs.html文件，写一点东西进去，后面被nginx用到
	os.Create("./test_ipvs.html")
	str := "Welcome to nginx"
	os.WriteFile("./test_ipvs.html", []byte(str), fs.FileMode(os.O_TRUNC))

	// 读取yaml文件，获取Pod和Service的配置信息
	var pod1 protocol.Pod
	yamlParse.YAMLParse(&pod1.Config, "../../../assets/test_ipvs/test_ipvs_pod1.yaml")
	pod1.Config.Metadata.UID = IPVS_TEST_UID_PREFIX + "1"

	var svc2 protocol.ServiceType
	yamlParse.YAMLParse(&svc2.Config, "../../../assets/test_ipvs/test_ipvs_service2_nodePort.yaml")
	svc2.Config.Metadata.UID = IPVS_TEST_UID_PREFIX + "4"

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
				ep := protocol.Endpoint{
					PodUID: pod1.Config.Metadata.UID,
					IP:     pod1_status.Status.IP,
					Port:   int(port.ContainerPort),
				}
				svc2.Status.Endpoints = append(svc2.Status.Endpoints, ep)
			}
		}
	}

	// 2. 创建一个service，手动指定ClusterIP，它NodePort在yaml文件中指定
	ops := NewIpvsOps(constant.CLUSTER_CIDR_DEFAULT)
	ops.Init()
	defer ops.Close()
	data, _ = yaml.Marshal(&svc2)
	fmt.Printf("Svc status: %s", string(data))

	svc2.Config.Spec.ClusterIP = "222.123.0.0"

	ops.AddService(&svc2)

	// 3. 在宿主机上使用本地网卡的IP地址，来curl这个service的nodePort；注意不要使用localhost，这是ipvs的一个小缺点（但NodePort本身就是用来给外部流量访问的）
	nodeIP, _ := net_util.GetNodeIP()
	// 我们只设置了1个NodePort暴露，故直接取它即可；如果有多个，需要遍历
	svc2_nodePort_addr_str := nodeIP + ":" + fmt.Sprint(svc2.Config.Spec.Ports[0].NodePort)
	output, err := exec.Command("curl", "--max-time", "3", svc2_nodePort_addr_str).Output()
	if !strings.Contains(string(output), "Welcome") {
		t.Fatalf("Error curl NodePort Service addr "+svc2_nodePort_addr_str+" from Host Machine, reason: %s", err)
	}

	// 4. 删除这个服务，后续应该不能访问到它！
	ops.DelService(&svc2)
	output, err = exec.Command("curl", "--max-time", "3", svc2_nodePort_addr_str).Output()
	// 应该不成功！
	if string(output) != "" {
		t.Fatalf("Delete NodePort Service addr "+svc2_nodePort_addr_str+", but still can curl it, reason: %s", err)
	}

	ClearAllPodsAndRules()
}
