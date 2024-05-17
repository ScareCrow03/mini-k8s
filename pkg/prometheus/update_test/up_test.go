package update_process_test

import (
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/prometheus/prmts_ops"
	"mini-k8s/pkg/protocol"
	rtm "mini-k8s/pkg/remoteRuntime/runtime"
	yamlParse "mini-k8s/pkg/utils/yaml"
	"testing"
	"time"
)

const PROMETHEUS_TEST_UID_PREFIX = "mini-k8s_test_PROMETHEUS-uid"

func ClearAll() {
	rtm_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer rtm_service.Close()

	var pod1 protocol.Pod
	yamlParse.YAMLParse(&pod1.Config, "../../../../assets/test_prometheus/test_prometheus_pod1.yaml")
	pod1.Config.Metadata.UID = PROMETHEUS_TEST_UID_PREFIX + "1"
	rtm_service.RemovePodSandBox(&pod1)
}

func TestMain(m *testing.M) {
	m.Run()
}

// 运行这个文件前，请保证Prometheus已经配置好，以及本地存在pod_metrics_image包定义的镜像
func TestUpdateProcess(t *testing.T) {
	ClearAll()

	rtm_service := rtm.NewRemoteRuntimeService(5 * time.Minute)
	defer rtm_service.Close()

	var pod1 protocol.Pod
	yamlParse.YAMLParse(&pod1.Config, "../../../../assets/test_prometheus/test_prometheus_pod1.yaml")
	pod1.Config.Metadata.UID = PROMETHEUS_TEST_UID_PREFIX + "1"

	rtm_service.CreatePod(&pod1)
	rtm_service.StartPod(&pod1)

	// 以下把这个Pod的端口暴露逻辑写进prometheus配置文件
	// 首先读取原配置
	cfg := prmts_ops.GetPrometheusConfigFromFile(constant.PrometheusConfigPath)

	// 然后从pods数组中筛选出需要暴露的Pod
	pods := []protocol.Pod{pod1}
	podsJobsName2Endpoints := prmts_ops.SelectPodsNeedExposeMetrics(pods)

	// 预先在本地运行一个NodeExporter，监控主机的资源，然后简单注册一下ip:port即可
	// 这个函数待完善成对一个Node结构体数组操作
	nodesJobsName2Endpoints := make(map[string][]string)
	nodesJobsName2Endpoints["local_node_exporter"] = []string{constant.MasterIp + ":9100"}

	// 结合上述两个map
	result := prmts_ops.MergeJobsName2Endpoints(podsJobsName2Endpoints, nodesJobsName2Endpoints)

	// 同步配置
	prmts_ops.SyncPrometheusConfig(&cfg, result)
	// 写回Prometheus配置文件，并热更新
	prmts_ops.ApplyPrometheusConfigToFile(cfg, constant.PrometheusConfigPath, constant.PrometheusReloadUrl)
}
