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
	case "dns":
		handleCreateDns(filePath)
	default:
		fmt.Println("unsupported object type:", objectType)
	}
	return nil
}

func handleCreatePod(filePath string) error {
	var pod1 protocol.Pod
	yaml.YAMLParse(&pod1.Config, filePath)
	req, err := json.Marshal(pod1.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createPodFromFile", req)
	return nil
}

func handleCreateService(filePath string) error {
	//TODO: 完成对service的创建
	fmt.Println("create service from file:", filePath)
	var service protocol.ServiceType
	yaml.YAMLParse(&service, filePath)
	req, err := json.Marshal(service)
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
	httputils.Post(constant.HttpPreffix+"/createDnsFromFile", req)
	return nil
}

func init() {
	createCmd.Flags().StringP("file", "f", "", "the file to create the resource from")
	createCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(createCmd)
}
