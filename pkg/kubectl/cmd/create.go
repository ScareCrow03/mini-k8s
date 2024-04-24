package cmd

import (
	"fmt"
	"mini-k8s/pkg/httputils"

	// "net/http"

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
	// 创建一个json格式的请求体，名字为filepath，然后发送一个post请求
	requestBody := make(map[string]interface{})
	requestBody["filepath"] = filePath
	httputils.Post("http://localhost:8080/createFromFile", requestBody)
	return nil
}

func init() {
	createCmd.Flags().StringP("file", "f", "", "the file to create the resource from")
	createCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(createCmd)
}
