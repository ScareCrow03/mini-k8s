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

// type FunctionControllerInterface interface {
// 	Run()
// }

type FunctionController struct {
	cache map[string]protocol.Function
}

func NewFunctionController() *FunctionController {
	return &FunctionController{
		cache: make(map[string]protocol.Function),
	}
}

func (fc *FunctionController) CheckFunction() {
	//向apiserver发送查询function请求，并与本地cache进行对比
	//如果apiserver返回的function与本地cache不一致，则更新本地cache
	//如果apiserver返回的function在本地cache中不存在，则添加到本地cache
	//如果本地cache中存在，但是apiserver返回的function不存在，则删除本地cache中的function
	fmt.Println("CheckAllFunction")
	jsonstr, err := json.Marshal("function")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", jsonstr)
	var functions []protocol.Function
	err = json.Unmarshal(resp, &functions)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cur := make(map[string]bool)
	for _, f := range functions {
		cur[f.Metadata.Namespace+"/"+f.Metadata.Name] = true
		if _, ok := fc.cache[f.Metadata.Namespace+"/"+f.Metadata.Name]; !ok {
			fmt.Printf("CreateFunction %s %s\n", f.Metadata.Namespace, f.Metadata.Name)
			fc.cache[f.Metadata.Namespace+"/"+f.Metadata.Name] = f
			fc.CreateFunction(f)
			fc.CreateReplica(f)
		}
	}
	//删除本地cache中不存在的function
	for k, f := range fc.cache {
		fmt.Println(k)
		if _, ok := cur[f.Metadata.Namespace+"/"+f.Metadata.Name]; !ok {
			fmt.Println("DeleteFunction", f.Metadata.Namespace, f.Metadata.Name)
			fc.DeleteFunction(f)
			delete(fc.cache, k)
		}
	}
}

func (fc *FunctionController) DeleteFunction(f protocol.Function) {
	//删除function对应的docker镜像
	var cli *client.Client
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer cli.Close()
	//直接删除对应的挂载卷
	mountpath := constant.WorkDir + "/assets/registry/docker/registry/v2/repositories/baseserver/" + f.Metadata.Namespace + "/" + f.Metadata.Name
	err = os.RemoveAll(mountpath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//删除function对应的replicaset
	fc.DeleteReplica(f)
}

func (fc *FunctionController) CreateFunction(f protocol.Function) {
	FunctionFilePath := constant.WorkDir + "/assets/" + f.Metadata.Namespace + "/" + f.Metadata.Name
	//首先创建docker镜像，然后将其推送到docker registry
	//创建文件夹,如果存在则删除后重建
	err := os.RemoveAll(FunctionFilePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = os.MkdirAll(FunctionFilePath, 0777)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//先检查zip文件是否存在，如果存在，则删除
	err = os.RemoveAll(FunctionFilePath + "/function.zip")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = os.RemoveAll(FunctionFilePath + "/function")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//根据function中的file字段，将其解压
	err = os.WriteFile(FunctionFilePath+"/function.zip", f.Spec.UserUploadFile, 0777)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//解压zip文件
	//将解压后的文件放入新文件夹
	z := archiver.NewZip()
	z.OverwriteExisting = true
	err = z.Unarchive(FunctionFilePath+"/function.zip", FunctionFilePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//删除压缩包
	err = os.Remove(FunctionFilePath + "/function.zip")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//创建dockerfile
	os.Remove(FunctionFilePath + "/Dockerfile")
	dockerfile, err := os.Create(FunctionFilePath + "/Dockerfile")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer dockerfile.Close()
	funcpath := f.Metadata.Name + "/"
	dockerfile.WriteString("FROM " + constant.BaseImage + ":latest\n")
	dockerfile.WriteString("COPY " + funcpath + " /app\n")
	dockerfile.WriteString("EXPOSE 10000\n")

	//构建docker上下文，需要将依赖文件打包成tar格式
	z2 := archiver.NewTar()
	z2.OverwriteExisting = true
	err = z2.Archive([]string{FunctionFilePath}, FunctionFilePath+"/function.tar")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	ctx, err := os.Open(FunctionFilePath + "/function.tar")
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
		Dockerfile: f.Metadata.Name + "/Dockerfile",
		Tags:       []string{constant.BaseImage + "/" + f.Metadata.Namespace + "/" + f.Metadata.Name},
		Context:    ctx,
		Remove:     true,
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
	//推送docker镜像到docker registry
	// jsonauth, err := json.Marshal(constant.Auth)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// fmt.Println(string(jsonauth))
	authEncoded := base64.StdEncoding.EncodeToString([]byte(constant.AuthCode))
	fmt.Println(authEncoded)
	resp2, err := cli.ImagePush(context.Background(), constant.BaseImage+"/"+f.Metadata.Namespace+"/"+f.Metadata.Name+":latest", image.PushOptions{
		RegistryAuth: authEncoded,
		All:          false,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp2.Close()
	io.Copy(os.Stdout, resp2)
	//删除新文件夹
	err = os.RemoveAll(FunctionFilePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}

func (fc *FunctionController) DeleteReplica(f protocol.Function) {
	var replica protocol.ReplicasetConfig
	replica.ApiVersion = "v1"
	replica.Kind = "Replicaset"
	replica.Metadata.Namespace = f.Metadata.Namespace
	replica.Metadata.Name = f.Metadata.Name
	replica.Spec.Replicas = 0
	replica.Spec.Selector.MatchLabels = make(map[string]string)
	replica.Spec.Template.Metadata.Labels = make(map[string]string)
	replica.Spec.Selector.MatchLabels["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name
	replica.Spec.Template.Metadata.Labels["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name
	replica.Spec.Template.Spec.Containers = make([]protocol.ContainerConfig, 1)
	replica.Spec.Template.Spec.Containers[0].Name = f.Metadata.Name
	replica.Spec.Template.Spec.Containers[0].Image = constant.BaseImage + "/" + f.Metadata.Namespace + "/" + f.Metadata.Name + ":latest"
	replica.Spec.Template.Spec.Containers[0].Ports = make([]protocol.CtrPortBindingType, 1)
	replica.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = 10000
	jsonstr, err := json.Marshal(replica)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	httputils.Post(constant.HttpPreffix+"/deleteReplicasetFromFile", jsonstr)
}

func (fc *FunctionController) CreateReplica(f protocol.Function) {
	var replica protocol.ReplicasetConfig
	replica.ApiVersion = "v1"
	replica.Kind = "Replicaset"
	replica.Metadata.Namespace = f.Metadata.Namespace
	replica.Metadata.Name = f.Metadata.Name
	replica.Spec.Replicas = 0
	replica.Spec.Selector.MatchLabels = make(map[string]string)
	replica.Spec.Template.Metadata.Labels = make(map[string]string)
	replica.Spec.Selector.MatchLabels["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name
	replica.Spec.Template.Metadata.Labels["FunctionMetadata"] = f.Metadata.Namespace + "/" + f.Metadata.Name
	replica.Spec.Template.Spec.Containers = make([]protocol.ContainerConfig, 1)
	replica.Spec.Template.Spec.Containers[0].Name = f.Metadata.Name
	replica.Spec.Template.Spec.Containers[0].Image = constant.BaseImage + "/" + f.Metadata.Namespace + "/" + f.Metadata.Name + ":latest"
	replica.Spec.Template.Spec.Containers[0].Ports = make([]protocol.CtrPortBindingType, 1)
	replica.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = 10000
	jsonstr, err := json.Marshal(replica)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	httputils.Post(constant.HttpPreffix+"/createReplicasetFromFile", jsonstr)
}

func (fc *FunctionController) Run() {
	//每10s进行一次routine操作
	fmt.Println("FunctionController Run")
	fc = NewFunctionController()
	ticker := time.NewTicker(10 * time.Second)
	// defer ticker.Stop()
	// 开启一个goroutine执行轮询操作
	go func() {
		for {
			select {
			case <-ticker.C:
				fc.CheckFunction()
			}
		}
	}()
}
