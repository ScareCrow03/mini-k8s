package cmd

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	kubelet2 "mini-k8s/pkg/kubelet"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	// 建议转成小写
	object = strings.ToLower(object)

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
			fmt.Println(p.Status.IP, p.Config.NodeName, p.Status.Phase, p.Status.Runtime, p.Status.UpdateTime)
		}
	} else if object == "service" {
		var services []protocol.ServiceType
		err := json.Unmarshal(resp, &services)
		if err != nil {
			panic(err)
		}
		for _, s := range services {
			data, _ := yaml.Marshal(s)
			fmt.Println(string(data))
		}
	} else if object == "dns" {
		var dnss []protocol.Dns
		err := json.Unmarshal(resp, &dnss)
		if err != nil {
			panic(err)
		}
		for _, d := range dnss {
			fmt.Println(d.Metadata.Name, d.Spec.Host)
		}
	} else if object == "replicaset" {
		var rs []protocol.ReplicasetType
		err := json.Unmarshal(resp, &rs)
		if err != nil {
			panic(err)
		}
		for _, r := range rs {
			data, _ := yaml.Marshal(r)
			fmt.Println(string(data))
		}
	} else if object == "hpa" {
		var hpas []protocol.HPAType
		err := json.Unmarshal(resp, &hpas)
		if err != nil {
			logger.KError("unmarshal hpas error %v", hpas)
			logger.KError("unmarshal hpas error %s", err)
		}

		for _, h := range hpas {
			data, _ := yaml.Marshal(h)
			fmt.Println(string(data))
		}

	} else if object == "node" {
		var nodes []kubelet2.Kubelet
		err := json.Unmarshal(resp, &nodes)
		if err != nil {
			logger.KError("unmarshal nodes error %v", nodes)
		}

		for _, n := range nodes {
			data, _ := yaml.Marshal(n)
			fmt.Println(string(data))
		}
	} else if object == "job" {
		var jobs []protocol.Job
		err := json.Unmarshal(resp, &jobs)
		if err != nil {
			logger.KError("unmarshal jobs error %v", jobs)
		}
		for _, j := range jobs {
			// data, _ := yaml.Marshal(j)
			name := j.Metadata.Name
			namespace := j.Metadata.Namespace
			starttime := j.Status.StartTime
			status := j.Status.JobState
			fmt.Printf("%s, %s, %s, %s ", name, namespace, starttime, status)
			if status == "Finished" {
				fmt.Printf(j.Status.OutputFileContent)
			} else if status == "Error" {
				fmt.Printf(j.Status.ErrorFileContent)
			}
			fmt.Printf("\n")
		}
	} else if object == "pv" {
		var pvs []protocol.PersistentVolume
		err := json.Unmarshal(resp, &pvs)
		if err != nil {
			panic(err)
		}
		for _, p := range pvs {
			data, _ := yaml.Marshal(p)
			fmt.Println(string(data))
		}
	} else if object == "pvc" {
		var pvcs []protocol.PersistentVolumeClaim
		err := json.Unmarshal(resp, &pvcs)
		if err != nil {
			panic(err)
		}
		for _, p := range pvcs {
			data, _ := yaml.Marshal(p)
			fmt.Println(string(data))
		}
	} else {
		// 认为这里是在获取用户自定义的资源
		var crs []protocol.CRType
		err := json.Unmarshal(resp, &crs)
		if err != nil {
			logger.KError("unmarshal crs error %v", crs)
		}

		for _, cr := range crs {
			data, _ := yaml.Marshal(cr)
			fmt.Println(string(data))
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
