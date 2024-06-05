package proxy_server

import (
	"io/fs"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/kubeproxy/ipvs_ops"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	yamlParse "mini-k8s/pkg/utils/yaml"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rtm_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer rtm_service.Close()
	ops := ipvs_ops.NewIpvsOps(constant.CLUSTER_CIDR_DEFAULT)
	defer ops.Close()
	ops.Clear()

	var pod1 protocol.Pod
	yamlParse.YAMLParse(&pod1.Config, "../../../assets/test_ipvs/test_ipvs_pod1.yaml")
	pod1.Config.Metadata.UID = PROXY_SERVER_TEST_UID_PREFIX + "1"
	rtm_service.RemovePodSandBox(&pod1)

	var pod2 protocol.Pod
	yamlParse.YAMLParse(&pod2.Config, "../../../assets/test_ipvs/test_ipvs_curl_pod.yaml")
	pod2.Config.Metadata.UID = PROXY_SERVER_TEST_UID_PREFIX + "3"
	rtm_service.RemovePodSandBox(&pod2)

	m.Run()

	rtm_service.RemovePodSandBox(&pod1)
	rtm_service.RemovePodSandBox(&pod2)
	ops.Clear()

}

const PROXY_SERVER_TEST_UID_PREFIX = "mini-k8s_test_proxy_server-uid"

func TestProxyServer(t *testing.T) {
	os.Create("/tmp/test_ipvs.html")
	str := "Welcome to nginx"
	os.WriteFile("/tmp/test_ipvs.html", []byte(str), fs.FileMode(os.O_TRUNC))

	rtm_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer rtm_service.Close()

	var pod1 protocol.Pod
	yamlParse.YAMLParse(&pod1.Config, "../../../assets/test_ipvs/test_ipvs_pod1.yaml")
	pod1.Config.Metadata.UID = PROXY_SERVER_TEST_UID_PREFIX + "1"

	var pod2 protocol.Pod
	yamlParse.YAMLParse(&pod2.Config, "../../../assets/test_ipvs/test_ipvs_curl_pod.yaml")
	pod2.Config.Metadata.UID = PROXY_SERVER_TEST_UID_PREFIX + "3"

	rtm_service.CreatePod(&pod1)
	rtm_service.CreatePod(&pod2)

	rtm_service.StartPod(&pod1)
	rtm_service.StartPod(&pod2)

	pod1.Status.Phase = constant.PodPhaseRunning
	pod2.Status.Phase = constant.PodPhaseRunning

	ps := NewProxyServer(constant.CLUSTER_CIDR_DEFAULT)
	ps.PodMap[pod1.Config.Metadata.UID] = &pod1
	ps.PodMap[pod2.Config.Metadata.UID] = &pod2

	var pod2_curl_id string
	for _, container := range pod2.Config.Spec.Containers {
		// 先找到本pod的curl容器
		if strings.Contains(container.Image, "curl") {
			pod2_curl_id = container.UID
		}
	}

	var svc1 protocol.ServiceType
	yamlParse.YAMLParse(&svc1.Config, "../../../assets/test_ipvs/test_ipvs_service1.yaml")
	svc1.Config.Metadata.UID = PROXY_SERVER_TEST_UID_PREFIX + "2"
	svc1.Config.Spec.ClusterIP = "222.111.0.0"
	svc1_clusterIP_addr_str := svc1.Config.Spec.ClusterIP + ":80"
	ps.OnServiceAdd(&svc1)

	// 现在这个svc1应该正确管理到上述pod1
	if len(ps.ServiceMap[svc1.Config.Metadata.UID].Status.Endpoints) != 1 {
		t.Errorf("svc1 should have 1 endpoint, but got %d", len(ps.ServiceMap[svc1.Config.Metadata.UID].Status.Endpoints))
	}

	// 而且可以从pod2访问到service1这个endpoints:pod1的nginx服务
	output_str, err := rtm_service.ExecContainer(pod2_curl_id, []string{"curl", "--max-time", "3", svc1_clusterIP_addr_str})

	if !strings.Contains(output_str, "Welcome") {
		t.Errorf("Error curl ClusterIP Service addr "+svc1.Config.Spec.ClusterIP+":80 from the other pod, reason: %s", err)
	}

	// 删除上述pod1的管理，这个方法内部直接调用了状态同步方法！
	ps.OnPodDelete(&pod1)

	// 现在svc1应该所有pod都不管理！注意内部发生了改动，并不是同一个对象了！
	if len(ps.ServiceMap[svc1.Config.Metadata.UID].Status.Endpoints) != 0 {
		t.Errorf("svc1 should have 0 endpoint, but got %d", len(ps.ServiceMap[svc1.Config.Metadata.UID].Status.Endpoints))
	}

	// 此时pod2不能访问到service1这个endpoints:pod1的nginx服务
	output_str, err = rtm_service.ExecContainer(pod2_curl_id, []string{"curl", "--max-time", "3", svc1_clusterIP_addr_str})

	if strings.Contains(output_str, "Welcome") {
		t.Errorf("Error curl ClusterIP Service addr "+svc1.Config.Spec.ClusterIP+":80 from the other pod, reason: %s", err)
	}

	// 重新添加回管理，也能自动更新状态！
	ps.OnPodAdd(&pod1)
	if len(ps.ServiceMap[svc1.Config.Metadata.UID].Status.Endpoints) != 1 {
		t.Errorf("svc1 should have 1 endpoint, but got %d", len(ps.ServiceMap[svc1.Config.Metadata.UID].Status.Endpoints))
	}

	// 现在pod2可以访问到service1这个endpoints:pod1的nginx服务
	output_str, err = rtm_service.ExecContainer(pod2_curl_id, []string{"curl", "--max-time", "3", svc1_clusterIP_addr_str})
	if !strings.Contains(output_str, "Welcome") {
		t.Errorf("Error curl ClusterIP Service addr "+svc1.Config.Spec.ClusterIP+":80 from other pod, reason: %s", err)
	}

	// 释放资源
	rtm_service.RemovePodSandBox(&pod1)
	rtm_service.RemovePodSandBox(&pod2)
	ps.IpvsOps.Clear()
}
