package cmd

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/protocol"

	"github.com/spf13/cobra"
)

// 定义 `kubectl get` 命令
// 用法：kubectl get [object]
// 第二种用法：kubectl get [object] [object-name]
// object包括：pod, service, dns, node, job, deployment, job,待补充
var getCmd = &cobra.Command{
	Use:   "get (-f [file] | TYPE [NAME])",
	Short: "Get objects from the cluster",
	Long:  "Get objects from the cluster. Supported objects: pods, services, etc.",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		if len(args) == 0 && file != "" {
			return getObjectByFile(file)
		}

		if file != "" && len(args) != 0 {
			cmd.Usage()
			return nil
		}

		if len(args) == 1 {
			object := args[0]
			return getObjectByType(object)
		}

		if len(args) == 2 {
			object := args[0]
			objectName := args[1]
			return getObjectByTypeAndName(object, objectName)
		}

		cmd.Usage()
		return nil
	},
}

func getObjectByType(object string) error {
	// 在这里实现根据资源类型获取资源的逻辑
	fmt.Printf("get object by type: %s\n", object)

	// 创建一个json格式的请求体，名字为object，然后发送一个post请求
	req, _ := json.Marshal(object)
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	if object == "pod" {
		var pods []protocol.Pod
		err := json.Unmarshal(resp, &pods)
		if err != nil {
			panic(err)
		}
		for _, p := range pods {
			fmt.Println(p.Config.Metadata.Name, p.Config.Metadata.Namespace)
			fmt.Println(p.Status.IP, p.Status.NodeName, p.Status.Phase, p.Status.Runtime, p.Status.UpdateTime)
		}
	}

	return nil
}

func getObjectByTypeAndName(object, objectName string) error {
	// 在这里实现根据资源类型和名称获取资源的逻辑
	fmt.Printf("get object: %s name is: %s", object, objectName)
	// 创建一个json格式的请求体，名字为object和objectName，然后发送一个post请求
	requestBody := make(map[string]interface{})
	requestBody["object"] = object
	requestBody["objectName"] = objectName
	// httputils.Post("http://localhost:8080/getByTypeAndName", requestBody)
	return nil
}

func getObjectByFile(filePath string) error {
	// 在这里实现从文件获取资源的逻辑
	fmt.Printf("get object from file: %s", filePath)
	return nil
}

func init() {
	// 为 getCmd 添加 -f 标志
	getCmd.Flags().StringP("file", "f", "", "get object from file")
	rootCmd.AddCommand(getCmd)
}
