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

// 将子命令添加到根命令中
func init() {
	// 	rootCmd.AddCommand(getCmd)
	// 	rootCmd.AddCommand(createCmd)
	// 	rootCmd.AddCommand(deleteCmd)
	// 	rootCmd.AddCommand(applyCmd)
	// 	rootCmd.AddCommand(describeCmd)
}

// Execute 是入口函数，用于执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
