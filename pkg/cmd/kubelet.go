package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// 定义根命令
var rootCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "kubectl is a CLI tool for Kubernetes",
	Long:  "kubectl is a command-line tool for managing Kubernetes clusters.",
}

// 定义 `kubectl get` 命令
var getCmd = &cobra.Command{
	Use:   "get [resource]",
	Short: "Get resources from the cluster",
	Long:  "Get resources from the cluster. Supported resources: pods, services, etc.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify a resource to get (e.g., pods, services).")
			return
		}

		resource := args[0]
		fmt.Printf("Getting %s from the cluster...\n", resource)
		// 这里可以添加从集群中获取资源的逻辑
		switch resource {
		case "pods":
			fmt.Println("Listing pods...")
		case "services":
			fmt.Println("Listing services...")
		default:
			fmt.Printf("Resource %s is not supported.\n", resource)
		}
	},
}

// 定义 `kubectl create` 命令
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources in the cluster",
	Long:  "Create resources in the cluster. Supported resources: pods, services, etc.",
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")
		if file == "" && len(args) != 1 {
			fmt.Println("Please specify a file to create the resource from.")
			return
		}

		if len(args) != 0 && file != "" {
			fmt.Println("Please specify either a file or arguments, not both.")
			return
		}

		if file == "" {
			fmt.Printf("Creating resource from args %v...\n", args)
			// 这里可以添加从参数中创建资源的逻辑
			return
		}

		fmt.Printf("Creating resource from file %s...\n", file)
		// 这里可以添加从文件中创建资源的逻辑
	},
}

// 定义 `kubectl delete` 命令
var deleteCmd = &cobra.Command{
	Use:   "delete [resource]",
	Short: "Delete resources from the cluster",
	Long:  "Delete resources from the cluster. Supported resources: pods, services, etc.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify a resource to delete (e.g., pod, service).")
			return
		}

		resource := args[0]
		fmt.Printf("Deleting %s from the cluster...\n", resource)
		// 这里可以添加从集群中删除资源的逻辑
	},
}

// 将子命令添加到根命令中
func init() {
	createCmd.Flags().StringP("file", "f", "", "the file to create the resource from")
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
}

// Execute 是入口函数，用于执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
