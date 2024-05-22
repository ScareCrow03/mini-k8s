package cmd

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/kubectl/kubeutils"
	"mini-k8s/pkg/protocol"

	yaml "mini-k8s/pkg/utils/yaml"

	"github.com/spf13/cobra"
)

// 定义 `kubectl create` 命令
// 用法：kubectl create -f [file]
var createCmd = &cobra.Command{
	Use:   "create -f [file]",
	Short: "Create resources in the cluster",
	Long:  "Create resources in the cluster. Supported resources: pods, services, etc.",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		if len(args) != 0 {
			cmd.Usage()
			return nil
		}
		return createFromFile(file)
	},
}

func createFromFile(filePath string) error {
	// 在这里实现从文件创建资源的逻辑
	fmt.Println("create resource from file:", filePath)
	objectType := kubeutils.GetTypeFromYAML(filePath)
	fmt.Println("object type:", objectType)
	switch objectType {
	case "Pod":
		handleCreatePod(filePath)
	case "Service":
		handleCreateService(filePath)
	case "Dns":
		handleCreateDns(filePath)
	case "Replicaset":
		handleCreateReplicaset(filePath)
	case "HPA":
	case "hpa":
	case "HorizontalPodAutoscaler":
		handleCreateHPA(filePath)

	case "Function":
		handleCreateFunction(filePath)
	default:
		handleCreateCR(filePath)
	}
	return nil
}

func handleCreatePod(filePath string) error {
	var pod protocol.Pod
	yaml.YAMLParse(&pod.Config, filePath)
	req, err := json.Marshal(pod.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	resp := httputils.Post(constant.HttpPreffix+"/createPodFromFile", req)
	var rss string
	json.Unmarshal(resp, &rss)
	fmt.Println(rss)
	return nil
}

func handleCreateService(filePath string) error {
	//TODO: 完成对service的创建
	fmt.Println("create service from file:", filePath)
	var service protocol.ServiceType
	yaml.YAMLParse(&service.Config, filePath)
	req, err := json.Marshal(service)
	fmt.Printf("req: %s\n", string(req))
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createServiceFromFile", req)
	return nil
}

func handleCreateDns(filePath string) error {
	//TODO: 完成对dns的创建
	fmt.Println("create dns from file:", filePath)
	var dns protocol.Dns
	yaml.YAMLParse(&dns, filePath)
	req, err := json.Marshal(dns)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	rep := httputils.Post(constant.HttpPreffix+"/createDnsFromFile", req)
	//响应体是{message: "create dns success"},将其反序列化
	var res map[string]interface{}
	json.Unmarshal(rep, &res)
	fmt.Println(res)
	return nil
}

func handleCreateReplicaset(filePath string) error {
	var rs protocol.ReplicasetType
	yaml.YAMLParse(&rs.Config, filePath)
	req, err := json.Marshal(rs.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createReplicasetFromFile", req)
	return nil
}

func handleCreateHPA(filePath string) error {
	var hpa protocol.HPAType
	yaml.YAMLParse(&hpa.Config, filePath)
	req, err := json.Marshal(hpa.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createHPAFromFile", req)
	return nil
}

func handleCreateCR(filePath string) error {
	var cr protocol.CRType
	yaml.YAMLParse(&cr, filePath)
	req, err := json.Marshal(cr)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createCRFromFile", req)
	return nil
}

func handleCreateFunction(filePath string) error {
	var function protocol.Function
	yaml.YAMLParse(&function, filePath)
	req, err := json.Marshal(function)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createFunctionFromFile", req)
	return nil
}

func init() {
	createCmd.Flags().StringP("file", "f", "", "the file to create the resource from")
	createCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(createCmd)
}
