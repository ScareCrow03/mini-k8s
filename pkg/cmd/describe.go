package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// 创建 describeCmd 命令
var describeCmd = &cobra.Command{
	Use:   "describe (-f FILENAME | TYPE [NAME_PREFIX])",
	Short: "describe resources in the cluster by file, resource type, or resource type and name",
	Long:  "describe resources in the cluster by file, resource type, or resource type and name. Supported resources: pods, services, etc.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取 -f 标志的文件路径
		filePath, _ := cmd.Flags().GetString("filename")
		// 根据文件描述资源
		if filePath != "" && len(args) == 0 {
			return describeFromFile(filePath)
		}
		// 根据资源类型、名称前缀或标签选择器描述资源
		if filePath == "" && len(args) >= 1 && len(args) <= 2 {
			resourceType := args[0]
			if len(args) == 2 {
				// 如果提供了两个参数，则是 TYPE/NAME 格式
				resourceName := args[1]
				return describeResourceByTypeAndName(resourceType, resourceName)
			} else {
				// 如果只提供了一个参数，则是 TYPE 格式
				return describeResourceByType(resourceType)
			}
		}
		// 无效输入返回错误
		cmd.Usage()
		return nil
	},
}

func init() {
	// 为 describeCmd 添加 -f 标志
	describeCmd.Flags().StringP("filename", "f", "", "get resource from file")
	// 将 describeCmd 添加到 rootCmd
	rootCmd.AddCommand(describeCmd)
}

// 从文件中描述资源的函数
func describeFromFile(filePath string) error {
	// 在这里实现从文件描述资源的逻辑
	fmt.Printf("describe resouce from file %s:", filePath)
	return nil
}

// 根据资源类型描述资源的函数
func describeResourceByType(resourceType string) error {
	// 在这里实现根据资源类型描述资源的逻辑
	fmt.Printf("describe resource from type %s", resourceType)
	return nil
}

// 根据资源类型和名称描述资源的函数
func describeResourceByTypeAndName(resourceType, resourceName string) error {
	// 在这里实现根据资源类型和名称描述资源的逻辑
	fmt.Printf("describe resouces: %s whose type is: %s\n", resourceName, resourceType)
	return nil
}
