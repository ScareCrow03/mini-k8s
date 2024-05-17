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

// 定义 `kubectl delete` 命令
// 用法：kubectl delete -f [file]
// 用法二：kubectl delete [resource] [resoucename]
var deleteCmd = &cobra.Command{
	Use:   "delete (-f [filename] | [resource] [resourcename])",
	Short: "Delete resources from the cluster by resource name",
	Long:  "Delete resources from the cluster. Supported resources: pods, services, etc.",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		if file != "" && len(args) == 0 {
			return deleteFromFile(file)
		}

		if file == "" && len(args) == 2 {
			resource := args[0]
			resourceName := args[1]
			return deleteResourceByName(resource, resourceName)
		}

		cmd.Usage()
		return nil
	},
}

func deleteFromFile(filePath string) error {
	// 在这里实现从文件删除资源的逻辑
	fmt.Printf("delete resource from file %s:", filePath)
	objectType := kubeutils.GetTypeFromYAML(filePath)
	fmt.Println("object type:", objectType)
	switch objectType {
	case "Pod":
		handleDeletePod(filePath)
	case "Service":
		handleDeleteService(filePath)
	case "Dns":
		handleDeleteDns(filePath)
	case "Replicaset":
		handleDeleteReplicaset(filePath)
	case "HPA":
	case "hpa":
	case "HorizontalPodAutoscaler":
		handleDeleteHPA(filePath)
	default:
		fmt.Println("unsupported object type:", objectType)
	}
	return nil
}

func handleDeletePod(filePath string) error {
	var pod protocol.Pod
	yaml.YAMLParse(&pod.Config, filePath)
	req, err := json.Marshal(pod.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/deletePodFromFile", req)
	return nil
}

func handleDeleteService(filePath string) error {
	fmt.Println("delete service from file:", filePath)
	var svc protocol.ServiceType
	yaml.YAMLParse(&svc.Config, filePath)
	req, err := json.Marshal(svc)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/deleteServiceFromFile", req)
	return nil
}

func handleDeleteDns(filePath string) error {
	var dns protocol.Dns
	yaml.YAMLParse(&dns, filePath)
	req, err := json.Marshal(dns)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/deleteDnsFromFile", req)
	return nil
}

func handleDeleteReplicaset(filePath string) error {
	fmt.Println("delete replicaset from file:", filePath)
	var rs protocol.ReplicasetType
	yaml.YAMLParse(&rs.Config, filePath)
	req, err := json.Marshal(rs.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/deleteReplicasetFromFile", req)
	return nil
}

func handleDeleteHPA(filePath string) error {
	fmt.Println("delete hpa from file:", filePath)
	var hpa protocol.HPAType
	yaml.YAMLParse(&hpa.Config, filePath)
	req, err := json.Marshal(hpa.Config)
	fmt.Println(string(req))

	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/deleteHPAFromFile", req)
	return nil
}

func deleteResourceByName(resource, resourceName string) error {
	// 在这里实现根据资源类型和名称删除资源的逻辑
	fmt.Printf("delete resource %s %s:", resource, resourceName)
	return nil
}

func init() {
	deleteCmd.Flags().StringP("file", "f", "", "the file to create the resource from")
	rootCmd.AddCommand(deleteCmd)
}
