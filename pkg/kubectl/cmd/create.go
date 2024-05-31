package cmd

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/kubectl/kubeutils"
	"mini-k8s/pkg/protocol"
	"os"
	"path/filepath"

	yaml "mini-k8s/pkg/utils/yaml"

	"github.com/mholt/archiver"
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
	objectType := kubeutils.GetTypeFromYAML(filePath)
	fmt.Println("object type:", objectType)
	switch objectType {
	case "Pod":
		handleCreatePod(filePath)
	case "Service":
		handleCreateService(filePath)
	case "Dns":
		handleCreateDns(filePath)
	case "Replicaset":
		handleCreateReplicaset(filePath)
	case "HPA":
	case "hpa":
	case "HorizontalPodAutoscaler":
		handleCreateHPA(filePath)

	case "Function":
		handleCreateFunction(filePath)
	case "Job":
		handleCreateJob(filePath)
	default:
		handleCreateCR(filePath)
	}
	return nil
}

func handleCreatePod(filePath string) error {
	var pod protocol.Pod
	yaml.YAMLParse(&pod.Config, filePath)
	req, err := json.Marshal(pod.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	resp := httputils.Post(constant.HttpPreffix+"/createPodFromFile", req)
	var rss string
	json.Unmarshal(resp, &rss)
	fmt.Println(rss)
	return nil
}

func handleCreateService(filePath string) error {
	//TODO: 完成对service的创建
	fmt.Println("create service from file:", filePath)
	var service protocol.ServiceType
	yaml.YAMLParse(&service.Config, filePath)
	req, err := json.Marshal(service)
	fmt.Printf("req: %s\n", string(req))
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createServiceFromFile", req)
	return nil
}

func handleCreateDns(filePath string) error {
	//TODO: 完成对dns的创建
	fmt.Println("create dns from file:", filePath)
	var dns protocol.Dns
	yaml.YAMLParse(&dns, filePath)
	req, err := json.Marshal(dns)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	rep := httputils.Post(constant.HttpPreffix+"/createDnsFromFile", req)
	//响应体是{message: "create dns success"},将其反序列化
	var res map[string]interface{}
	json.Unmarshal(rep, &res)
	fmt.Println(res)
	return nil
}

func handleCreateReplicaset(filePath string) error {
	var rs protocol.ReplicasetType
	yaml.YAMLParse(&rs.Config, filePath)
	req, err := json.Marshal(rs.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createReplicasetFromFile", req)
	return nil
}

func handleCreateHPA(filePath string) error {
	var hpa protocol.HPAType
	yaml.YAMLParse(&hpa.Config, filePath)
	req, err := json.Marshal(hpa.Config)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createHPAFromFile", req)
	return nil
}

func handleCreateCR(filePath string) error {
	var cr protocol.CRType
	yaml.YAMLParse(&cr, filePath)
	req, err := json.Marshal(cr)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createCRFromFile", req)
	return nil
}

func handleCreateFunction(filePath string) error {
	var function protocol.Function
	yaml.YAMLParse(&function, filePath)

	//找到对应的目录，将其打包成zip文件
	_, err := os.Stat(function.Spec.UserUploadPath)
	fmt.Println("file path:", function.Spec.UserUploadPath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	//压缩成zip文件
	z := archiver.NewZip()
	z.OverwriteExisting = true
	files, err := filepath.Glob(filepath.Join(function.Spec.UserUploadPath, "*"))
	if err != nil {
		return err
	}

	// 相对路径的压缩，将文件直接放在 ZIP 根目录
	relativeFiles := make([]string, len(files))
	for i, file := range files {
		relativeFiles[i] = filepath.Base(file)
	}
	err = z.Archive(files, function.Spec.UserUploadPath+".zip")
	if err != nil {
		return err
	}

	//将zip文件转化成byte
	file, err := os.ReadFile(function.Spec.UserUploadPath + ".zip")
	if err != nil {
		fmt.Println("read file failed")
		return err
	}

	function.Spec.UserUploadFile = file

	req, err := json.Marshal(function)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createFunctionFromFile", req)
	os.RemoveAll(function.Spec.UserUploadPath + ".zip")
	return nil
}

func handleCreateJob(filePath string) error {
	var job protocol.Job
	yaml.YAMLParse(&job, filePath)

	//找到对应的目录，将其打包成zip文件
	_, err := os.Stat(job.Spec.UploadPath)
	fmt.Println("file path:", job.Spec.UploadPath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	//压缩成zip文件
	z := archiver.NewZip()
	z.OverwriteExisting = true
	files, err := filepath.Glob(filepath.Join(job.Spec.UploadPath, "*"))
	if err != nil {
		return err
	}

	// 相对路径的压缩，将文件直接放在 ZIP 根目录
	relativeFiles := make([]string, len(files))
	for i, file := range files {
		relativeFiles[i] = filepath.Base(file)
	}
	err = z.Archive(files, job.Spec.UploadPath+".zip")
	if err != nil {
		return err
	}

	//将zip文件转化成byte
	file, err := os.ReadFile(job.Spec.UploadPath + ".zip")
	if err != nil {
		fmt.Println("read file failed")
		return err
	}

	job.Spec.UserUploadFile = file

	req, err := json.Marshal(job)
	if err != nil {
		fmt.Println("marshal request body failed")
		return err
	}
	httputils.Post(constant.HttpPreffix+"/createJobFromFile", req)
	os.RemoveAll(job.Spec.UploadPath + ".zip")
	return nil
}

func init() {
	createCmd.Flags().StringP("file", "f", "", "the file to create the resource from")
	createCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(createCmd)
}
