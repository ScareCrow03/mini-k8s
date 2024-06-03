package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/protocol"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/mholt/archiver"
)

type JobController struct {
	cache map[string]protocol.Job
}

func NewJobController() *JobController {
	return &JobController{
		cache: make(map[string]protocol.Job),
	}
}

func (fc *JobController) CheckJob() {
	fmt.Println("CheckJob")
	jsonstr, err := json.Marshal("job")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", jsonstr)
	var jobs []protocol.Job
	err = json.Unmarshal(resp, &jobs)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cur := make(map[string]bool)
	for _, job := range jobs {
		cur[job.Metadata.Namespace+"/"+job.Metadata.Name] = true
		if _, ok := fc.cache[job.Metadata.Namespace+"/"+job.Metadata.Name]; !ok {
			fmt.Printf("CreateJob %s %s\n", job.Metadata.Namespace, job.Metadata.Name)
			fc.cache[job.Metadata.Namespace+"/"+job.Metadata.Name] = job
			fc.CreateJob(job)
		}
	}
}

func (fc *JobController) CreateJob(job protocol.Job) {
	JobFilePath := constant.WorkDir + "/jobs/" + job.Metadata.Namespace + "/" + job.Metadata.Name
	// 创建job文件
	err := os.RemoveAll(JobFilePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = os.MkdirAll(JobFilePath, 0777)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = os.WriteFile(JobFilePath+"/job.zip", job.Spec.UserUploadFile, 0777)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	z := archiver.NewZip()
	z.OverwriteExisting = true
	err = z.Unarchive(JobFilePath+"/job.zip", JobFilePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 写slurm文件
	scriptanme := job.Metadata.Namespace + "_" + job.Metadata.Name + ".slurm"
	slurm, err := os.Create(JobFilePath + "/" + scriptanme)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer slurm.Close()
	slurm.WriteString("#!/bin/bash\n")
	slurm.WriteString("#SBATCH --job-name=" + job.Metadata.Namespace + "_" + job.Metadata.Name + "\n")
	slurm.WriteString("#SBATCH --output=" + job.Spec.OutputFile + "\n")
	slurm.WriteString("#SBATCH --error=" + job.Spec.ErrorFile + "\n")
	slurm.WriteString("#SBATCH --partition=" + job.Spec.Partition + "\n")
	slurm.WriteString("#SBATCH -n " + job.Spec.NTasks + "\n")
	slurm.WriteString("#SBATCH --ntasks-per-node=" + job.Spec.NTasksPerNode + "\n")
	slurm.WriteString("#SBATCH --gres=gpu:" + job.Spec.GPUNums + "\n")
	slurm.WriteString("#SBATCH --cpus-per-task=" + job.Spec.CPUPerTask + "\n")

	for _, v := range job.Spec.RunCommands {
		slurm.WriteString(v + "\n")
	}

	//构建dockerfile
	dockerfile, err := os.Create(JobFilePath + "/Dockerfile")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer dockerfile.Close()
	dockerfile.WriteString("FROM " + constant.JOB_SERVER_IMAGE + ":latest\n")
	dockerfile.WriteString("ENV OUTPUT_FILE " + job.Spec.OutputFile + "\n")
	dockerfile.WriteString("ENV ERROR_FILE " + job.Spec.ErrorFile + "\n")
	dockerfile.WriteString("ENV API_SERVER_IP " + constant.MasterIp + "\n")
	dockerfile.WriteString("ENV API_SERVER_PORT " + constant.API_SERVER_PORT + "\n")
	dockerfile.WriteString("ENV JOB_NAME " + job.Metadata.Name + "\n")
	dockerfile.WriteString("ENV JOB_NAMESPACE " + job.Metadata.Namespace + "\n")
	dockerfile.WriteString("COPY ./" + job.Metadata.Name + " /app/func\n")

	//构建docker上下文，需要将依赖文件打包成tar格式
	z2 := archiver.NewTar()
	z2.OverwriteExisting = true
	err = z2.Archive([]string{JobFilePath}, JobFilePath+"/job.tar")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	ctx, err := os.Open(JobFilePath + "/job.tar")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//构建docker镜像
	var cli *client.Client
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer cli.Close()

	resp, err := cli.ImageBuild(context.Background(), ctx, types.ImageBuildOptions{
		Dockerfile: job.Metadata.Name + "/Dockerfile",
		Tags:       []string{constant.JOB_SERVER_IMAGE + "/" + job.Metadata.Namespace + "/" + job.Metadata.Name},
		Context:    ctx,
		Remove:     true,
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)

	authEncoded := base64.StdEncoding.EncodeToString([]byte(constant.AuthCode))
	fmt.Println(authEncoded)
	resp2, err := cli.ImagePush(context.Background(), constant.JOB_SERVER_IMAGE+"/"+job.Metadata.Namespace+"/"+job.Metadata.Name+":latest", image.PushOptions{
		RegistryAuth: authEncoded,
		All:          false,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp2.Close()
	io.Copy(os.Stdout, resp2)

	//创建pod
	var pod protocol.Pod
	pod.Config.ApiVersion = "v1"
	pod.Config.Kind = "Pod"
	pod.Config.Metadata.Name = "job_pod_" + job.Metadata.Name
	pod.Config.Metadata.Namespace = job.Metadata.Namespace
	pod.Config.Spec.Containers = make([]protocol.ContainerConfig, 0)
	pod.Config.Spec.Containers = append(pod.Config.Spec.Containers, protocol.ContainerConfig{
		Name:  "job_container_" + job.Metadata.Name,
		Image: constant.JOB_SERVER_IMAGE + "/" + job.Metadata.Namespace + "/" + job.Metadata.Name,
	})
	pod.Config.Spec.RestartPolicy = "Onfailure"

	//发送创建pod请求
	podjson, err := json.Marshal(pod.Config)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	httputils.Post(constant.HttpPreffix+"/createPodFromFile", podjson)
}

func (fc *JobController) Start() {
	//每10s进行一次routine操作
	fmt.Println("JobController Run")
	fc = NewJobController()
	ticker := time.NewTicker(10 * time.Second)
	// defer ticker.Stop()
	// 开启一个goroutine执行轮询操作
	go func() {
		for {
			select {
			case <-ticker.C:
				fc.CheckJob()
			}
		}
	}()
}
