package controller

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/protocol"
	"os"
	"time"
)

type FunctionController interface {
	Run()
}

type functionController struct {
	cache map[string]protocol.Function
}

func NewFunctionController() FunctionController {
	return &functionController{
		cache: make(map[string]protocol.Function),
	}
}

func (fc *functionController) CheckFunction() {
	//向apiserver发送查询function请求，并与本地cache进行对比
	//如果apiserver返回的function与本地cache不一致，则更新本地cache
	//如果apiserver返回的function在本地cache中不存在，则添加到本地cache
	//如果本地cache中存在，但是apiserver返回的function不存在，则删除本地cache中的function
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
	for _, f := range functions {
		if _, ok := fc.cache[f.Metadata.Namespace+"/"+f.Metadata.Name]; !ok {
			fc.cache[f.Metadata.Namespace+"/"+f.Metadata.Name] = f
			fc.CreateFunction(f)
		}
	}
}

func (fc *functionController) CreateFunction(f protocol.Function) {
	filepath := constant.WorkDir + "/" + f.Metadata.Namespace + "/" + f.Metadata.UID + "/" + f.Spec.Source
	//首先创建docker镜像，然后将其推送到docker registry
	err := os.Mkdir(filepath, 0777)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//创建dockerfile
	dockerfile, err := os.Create(filepath + "/Dockerfile")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer dockerfile.Close()
	dockerfile.WriteString("FROM " + constant.BaseImage + "\n")
	dockerfile.WriteString("COPY . /app\n")
	dockerfile.WriteString("EXPOSE 10000\n")

}

func (fc *functionController) Run() {
	//每10s进行一次routine操作
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	// 开启一个goroutine执行轮询操作
	go func() {
		for {
			<-ticker.C // 当定时器触发时，执行下面的操作
			// 调用你想要执行的函数
			fc.CheckFunction()
		}
	}()
}
