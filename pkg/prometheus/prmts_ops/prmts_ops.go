package prmts_ops

import (
	"fmt"
	"mini-k8s/pkg/constant"
	kubelet2 "mini-k8s/pkg/kubelet"
	"mini-k8s/pkg/protocol"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/percona/promconfig"
	"gopkg.in/yaml.v3"
)

const (
	// 指定了这个字段的Pods，可以被纳入Prometheus的监控范围，请保证这里的Port是可以被外界访问的！
	// 可以指定1个或多个，要求分隔符为逗号，形如"2112,2113"
	PodsAnnotationsForMetrics = "prometheus.io/scrapePorts"
)

// 这是基于文件的服务发现；事实上还可以先搭一个consul注册中心，只需要在主程序中与consul注册自己、然后配置健康检查TTL，然后开一个协程定时向consul续约，这样只要这个进程退出，心跳中断一会，consul就可以自动删掉它；坏处是这个发心跳的逻辑必须在程序中额外显式写出来，不够优雅
// 还是建议手动用一个守护进程来拉取api-server的数据，然后更新配置文件即可，而不是将任务offload到每个服务本身
// 给定name->endpoints映射、其中name是字符串，endpoints是形如"127.0.0.1:2112"的串，同步到Prometheus配置结构体中
// 关键问题，update_process如何觉察到某个Pod的某个端口，提供了/metrics服务可供Prometheus拉取？（应该注意prometheus默认按http向某个进程<ip>:<port>拉取服务的路由都是/metrics，这个是可以事先保证的）
// 方法1. 在update_process中用http探针，访问所有Pod的每一个端口的/metircs服务，如果返回200，就认为这个Pod是一个Prometheus的target，可以添加到配置文件中，但这样实在太冗长了
// 方法2. 在Pod的yaml文件中，添加一个注解（注解本身就是用来传递信息的，它并不参与Pod的创建、筛选逻辑！），告诉update_process这个Pod是暴露了prometheus需要的/metrics服务的，但这仍然需要一个探针轮询本Pods的所有端口；所以建议要暴露/metrics服务的端口，也在注解中显式给出来，这样update_process的查找效率就非常高了！这要求用户稍微改动一下yaml文件、显式指定Prometheus需要的服务端口，而不能做到直接配置容器发指标后就完全无感知，但在我们的简单场景下值得的

// 这个函数用于生成规范的name，形如"pod-{namespace}/{name}"
func GetFormattedName(namespace string, name string, prefix string) string {
	return "minik8s-" + prefix + "-" + namespace + "/" + name
}

func GetPrometheusConfigFromFile(path string) promconfig.Config {
	// Prometheus配置文件路径
	configPath := constant.PrometheusConfigPath

	// 读取Prometheus配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Failed to read Prometheus configuration file: %s\n", err)
	}

	// 解析Prometheus配置文件
	var cfg promconfig.Config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Printf("Failed to unmarshal Prometheus configuration: %s\n", err)
	}

	return cfg
}

func ApplyPrometheusConfigToFile(cfg promconfig.Config, path string, reloadUrl string) {
	// 序列化Prometheus配置
	data, err := yaml.Marshal(cfg)
	if err != nil {
		fmt.Printf("Failed to marshal Prometheus configuration: %s\n", err)
	}

	// 写回Prometheus配置文件
	err = os.WriteFile(path, data, 0777)
	if err != nil {
		fmt.Printf("Failed to write Prometheus configuration file: %s\n", err)
	}

	if reloadUrl != "" {
		_, err = exec.Command("curl", "--max-time", "5", "-X", "POST", reloadUrl).Output()
		if err != nil {
			fmt.Printf("Failed to reload Prometheus configuration: %s\n", err)
		}
	}
}

// 除了prometheus自身的监控外，完全同步到给定的jobsName2Endpoints映射，认为一个jobname可以有多个endpoints（比如同一个ip的不同端口）
func SyncPrometheusConfig(cfg *promconfig.Config, jobsName2Endpoints map[string][]string) {
	// 创建一个新的ScrapeConfigs切片，用于存储新的jobs
	newScrapeConfigs := make([]*promconfig.ScrapeConfig, 0)

	// 保留Prometheus自身的默认监听job
	for _, scrapeConfig := range cfg.ScrapeConfigs {
		if scrapeConfig.JobName == "prometheus" {
			newScrapeConfigs = append(newScrapeConfigs, scrapeConfig)
		}
	}

	// 添加新的jobs端点
	for jobName, endpoints := range jobsName2Endpoints {
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
		newScrapeConfigs = append(newScrapeConfigs, newJob)
	}

	// 更新ScrapeConfigs
	cfg.ScrapeConfigs = newScrapeConfigs
}

func SelectNodesNeedExposeMetrics(nodes []kubelet2.Kubelet) map[string][]string {
	// 默认所有Nodes都需要监听运行在9100端口的NodeExporter
	jobsName2Endpoints := make(map[string][]string)
	for _, node := range nodes {
		jobName := GetFormattedName("", node.Config.Name, "node")
		// 可以考虑也添加一个探针逻辑，但默认NodeExporter已经运行在9100端口故不用继续管。
		endpoint := node.Config.NodeIP + ":9100"
		jobsName2Endpoints[jobName] = append(jobsName2Endpoints[jobName], endpoint)
	}
	return jobsName2Endpoints
}

func SelectPodsNeedExposeMetrics(pods []protocol.Pod) map[string][]string {
	// 遍历所有Pod，根据annotation字段找出需要暴露/metrics服务的Pod
	jobsName2Endpoints := make(map[string][]string)
	for _, pod := range pods {
		// // 遍历Pod的所有容器
		// // 显式在annotation中声明的方法
		// if pod.Config.Metadata.Annotations[PodsAnnotationsForMetrics] != "" && pod.Status.IP != "" {
		// 	// 如果指定了暴露这个Port供访问，而且它已经有PodIP，那么可以作为一个endpoints
		// 	portsStr := pod.Config.Metadata.Annotations[PodsAnnotationsForMetrics]
		// 	parts := strings.Split(portsStr, ",")
		// 	result := make([]int, len(parts)) // 建立一个port整型数组
		// 	for i, part := range parts {
		// 		port, err := strconv.Atoi(strings.TrimSpace(part))
		// 		if err != nil {
		// 			fmt.Printf("Failed to convert prometheus exposed port to int: %s\n", err)
		// 			continue
		// 		}
		// 		result[i] = port
		// 	}

		// 	// 把这个Pod暴露的若干端点提取出来，加入映射
		// 	jobName := GetFormattedName(pod.Config.Metadata.Namespace, pod.Config.Metadata.Name, "pod")
		// 	for _, port := range result {
		// 		endpoint := pod.Status.IP + ":" + strconv.Itoa(port)
		// 		jobsName2Endpoints[jobName] = append(jobsName2Endpoints[jobName], endpoint)
		// 	}
		// }

		// 探针式的方法，检查某个端口/metrics服务是否可用
		if pod.Status.IP != "" {
			// 如果Pod已经有PodIP，那么可以作为一个endpoints，那么依次探测这个Pod的所有端口，如果有/metrics服务，就加入映射
			jobName := GetFormattedName(pod.Config.Metadata.Namespace, pod.Config.Metadata.Name, "pod")
			for _, container := range pod.Config.Spec.Containers {
				for _, port := range container.Ports {
					if ProbeMetricsService(pod.Status.IP, fmt.Sprint(port.ContainerPort)) {
						endpoint := pod.Status.IP + ":" + fmt.Sprint(port.ContainerPort)
						jobsName2Endpoints[jobName] = append(jobsName2Endpoints[jobName], endpoint)
					}
				}
			}
		}
	}

	return jobsName2Endpoints
}

// 合并两个映射s
func MergeJobsName2Endpoints(map1 map[string][]string, map2 map[string][]string) map[string][]string {
	result := make(map[string][]string)
	for k, v := range map1 {
		result[k] = append(result[k], v...)
	}
	for k, v := range map2 {
		result[k] = append(result[k], v...)
	}

	return result
}

// 探测指定<ip>:<port>的/metrics服务是否可用
func ProbeMetricsService(ip string, port string) bool {
	url := fmt.Sprintf("http://%s:%s/metrics", ip, port)

	// 创建一个http客户端
	client := &http.Client{
		Timeout: 1 * time.Second, // 设置超时时间
	}

	// 发送一个GET请求
	resp, err := client.Get(url)
	if err != nil {
		// 没有提供这个服务
		return false
	}
	defer resp.Body.Close()

	// 检查HTTP响应的状态码
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Find a /metrics service at", url)
		return true
	} else {
		return false
	}
}
