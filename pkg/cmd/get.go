package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// 定义 `kubectl get` 命令
// 用法：kubectl get [resource]
// 第二种用法：kubectl get [resource] [resource-name]
// resouce包括：pods, services, dns, node, job, deployment, job,待补充
var getCmd = &cobra.Command{
	Use:   "get (-f [file] | TYPE [NAME])",
	Short: "Get resources from the cluster",
	Long:  "Get resources from the cluster. Supported resources: pods, services, etc.",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		if len(args) == 0 && file != "" {
			return getResourceByFile(file)
		}

		if file != "" && len(args) != 0 {
			cmd.Usage()
			return nil
		}

		if len(args) == 1 {
			resource := args[0]
			return getResourceByType(resource)
		}

		if len(args) == 2 {
			resource := args[0]
			resourceName := args[1]
			return getResourceByTypeAndName(resource, resourceName)
		}

		cmd.Usage()
		return nil
	},
}

func getResourceByType(resource string) error {
	// 在这里实现根据资源类型获取资源的逻辑
	fmt.Printf("get resource by type: %s", resource)
	return nil
}

func getResourceByTypeAndName(resource, resourceName string) error {
	// 在这里实现根据资源类型和名称获取资源的逻辑
	fmt.Printf("get resource: %s name is: %s", resource, resourceName)
	return nil
}

func getResourceByFile(filePath string) error {
	// 在这里实现从文件获取资源的逻辑
	fmt.Printf("get resource from file: %s", filePath)
	return nil
}

func init() {
	// 为 getCmd 添加 -f 标志
	getCmd.Flags().StringP("file", "f", "", "get resource from file")
	rootCmd.AddCommand(getCmd)
}
