package cmd

import (
	"fmt"
	"mini-k8s/pkg/httputils"

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
	// 创建一个json格式的请求体，名字为filepath，然后发送一个post请求
	requestBody := make(map[string]interface{})
	requestBody["filepath"] = filePath
	httputils.Post("http://localhost:8080/deleteFromFile", requestBody)
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
