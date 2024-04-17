package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// 定义 `kubectl apply` 命令
// 用法：kubectl apply -f [file]
var applyCmd = &cobra.Command{
	Use:   "apply -f [file]",
	Short: "Apply resources to the cluster",
	Long:  "Apply resources to the cluster. Supported resources: pods, services, etc.",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		if len(args) != 0 {
			cmd.Usage()
			return nil
		}
		return applyFromFile(file)
	},
}

func applyFromFile(filePath string) error {
	// 在这里实现从文件应用资源的逻辑
	fmt.Println("apply resource from file:", filePath)
	return nil
}

func init() {
	applyCmd.Flags().StringP("file", "f", "", "the file to apply the resource from")
	applyCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(applyCmd)
}
