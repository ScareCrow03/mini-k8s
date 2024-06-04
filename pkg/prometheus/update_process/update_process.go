package update_process

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	kubelet2 "mini-k8s/pkg/kubelet"
	"mini-k8s/pkg/prometheus/prmts_ops"
	"mini-k8s/pkg/protocol"
	"strings"
	"time"

	"github.com/percona/promconfig"
)

type UpdateProcess struct {
	PodsMap  map[string]*protocol.Pod // 新增一个map来存储每个Pod对象
	NodesMap map[string]*kubelet2.Kubelet
}

func (up *UpdateProcess) Start() {
	up.PodsMap = make(map[string]*protocol.Pod)
	up.NodesMap = make(map[string]*kubelet2.Kubelet)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	fmt.Printf("mini-k8s Prometheus update process init sync, it may take some time...\n")
	// 第一次的源文件请先去掉名称中包含minik8s的所有东西，然后再sync
	cfg := prmts_ops.GetPrometheusConfigFromFile(constant.PrometheusConfigPath)
	for i := 0; i < len(cfg.ScrapeConfigs); {
		if strings.Contains(cfg.ScrapeConfigs[i].JobName, "minik8s") {
			// 删掉这个job名称
			cfg.ScrapeConfigs = append(cfg.ScrapeConfigs[:i], cfg.ScrapeConfigs[i+1:]...)
		} else {
			i++
		}
	}

	up.SyncAllPodsAndNodes(&cfg)
	prmts_ops.ApplyPrometheusConfigToFile(cfg, constant.PrometheusConfigPath, constant.PrometheusReloadUrl)

	fmt.Printf("mini-k8s Prometheus update process init sync finished! \n")
	go func() {
		for {
			select {
			case <-ticker.C:
				{
					up.DoUpdate()
				}
			}
		}
	}()

	select {}
}

func (up *UpdateProcess) DoUpdate() {
	fmt.Println("do Prometheus update")
	cfg := prmts_ops.GetPrometheusConfigFromFile(constant.PrometheusConfigPath)
	up.SyncAllPodsAndNodes(&cfg)
	prmts_ops.ApplyPrometheusConfigToFile(cfg, constant.PrometheusConfigPath, constant.PrometheusReloadUrl)
}

func (up *UpdateProcess) SyncAllPodsAndNodes(cfg *promconfig.Config) {
	fmt.Println("SyncAllPodsAndNodes")
	req, _ := json.Marshal("pod")
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	var pods []protocol.Pod
	json.Unmarshal(resp, &pods)

	// 拉取所有nodes
	req2, _ := json.Marshal("node")
	resp2 := httputils.Post(constant.HttpPreffix+"/getObjectByType", req2)
	var nodes []kubelet2.Kubelet
	json.Unmarshal(resp2, &nodes)

	updatedPods := make(map[string]*protocol.Pod)
	updatedNodes := make(map[string]*kubelet2.Kubelet)

	var newCreatedPods []protocol.Pod
	var newCreatedNodes []kubelet2.Kubelet

	needRemoveJobsName := make(map[string]bool)

	// 查看哪些是新的pods
	for _, pod := range pods {
		podKey := pod.Config.Metadata.UID
		updatedPods[podKey] = &pod
		if _, ok := up.PodsMap[podKey]; !ok {
			if pod.Status.IP != "" { // 只能选取有IP的Pod加入到Map中，否则出现问题！
				newCreatedPods = append(newCreatedPods, pod)
				up.PodsMap[podKey] = &pod
			}
		}
	}

	// 查看哪些是新的nodes
	for _, node := range nodes {
		nodeKey := node.Config.Name // Node没有UID，直接都按名字来索引
		updatedNodes[nodeKey] = &node
		if _, ok := up.NodesMap[nodeKey]; !ok {
			newCreatedNodes = append(newCreatedNodes, node)
			up.NodesMap[nodeKey] = &node
		}
	}

	// 对于PodsMap中有但是在新的Pod列表中没有的，删除对应的job
	for podKey, pod := range up.PodsMap {
		if _, ok := updatedPods[podKey]; !ok {
			jobsName := prmts_ops.GetFormattedName(pod.Config.Metadata.Namespace, pod.Config.Metadata.Name, "pod") // 获取它的名字
			needRemoveJobsName[jobsName] = true
			delete(up.PodsMap, podKey)
		}
	}

	// 对于NodesMap中有但是在新的Node列表中没有的，删除对应的job
	for nodeKey, node := range up.NodesMap {
		if _, ok := updatedNodes[nodeKey]; !ok {
			jobsName := prmts_ops.GetFormattedName("", node.Config.Name, "node") // 获取它的名字
			needRemoveJobsName[jobsName] = true
			delete(up.NodesMap, nodeKey)
		}
	}

	if len(needRemoveJobsName) > 0 {
		fmt.Printf("needRemoveJobsName: %v\n", needRemoveJobsName)
		// 修改配置，先删掉已经不存在的Pod/Node的job，方法是按jobName遍历cfg里的eps数组，如果在这个能被删除的集合里，那么去掉它。
		for i := 0; i < len(cfg.ScrapeConfigs); {
			if _, ok := needRemoveJobsName[cfg.ScrapeConfigs[i].JobName]; ok {
				// 删掉这个job名称
				cfg.ScrapeConfigs = append(cfg.ScrapeConfigs[:i], cfg.ScrapeConfigs[i+1:]...)
			} else {
				i++
			}
		}
	}

	// 然后添加新的Pod的job，这里可以复用之前的代码，筛选出新Pods里暴露了/metrics的Pod
	// 注意，这里Pods的增量可能导致第一次获取Pods时，它的IP没有，这些Pods不能加入Map！
	podsJobsName2Endpoints := prmts_ops.SelectPodsNeedExposeMetrics(newCreatedPods)
	// 添加新的Node的job，也可以复用
	nodesJobsName2Endpoints := prmts_ops.SelectNodesNeedExposeMetrics(newCreatedNodes)

	// 合并两个map
	result := prmts_ops.MergeJobsName2Endpoints(podsJobsName2Endpoints, nodesJobsName2Endpoints)
	// 建立一个当前配置里的job名称的集合
	nowCfgNames := make(map[string]bool)
	for _, eps := range cfg.ScrapeConfigs {
		nowCfgNames[eps.JobName] = true
	}

	if len(result) > 0 {
		fmt.Printf("needCreatedJobsAndEps: %v\n", result)
		// 添加它们到cfg里
		for jobName, endpoints := range result {
			newJob := &promconfig.ScrapeConfig{
				JobName: jobName,
				ServiceDiscoveryConfig: promconfig.ServiceDiscoveryConfig{
					StaticConfigs: make([]*promconfig.Group, len(endpoints)),
				},
			}
			for i, endpoint := range endpoints {
				newJob.ServiceDiscoveryConfig.StaticConfigs[i] = &promconfig.Group{
					Targets: []string{endpoint},
				}
			}

			// 如果有这个jobName，那么不必添加
			if _, ok := nowCfgNames[jobName]; !ok {
				cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, newJob)
			}
		}
	}

}
