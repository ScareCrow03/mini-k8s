## cmd

一键删除所有容器

```bash
docker stop $(docker ps -aq) && docker rm $(docker ps -aq)
```



启动

```bash
go run pkg/apiserver/main/main.go

go run pkg/scheduler/main/main.go

go run pkg/kubelet/main/main.go
```



```
go run pkg/kubectl/main/main.go get pod

go run pkg/kubectl/main/main.go create -f assets/pod_create_test1.yaml

go run pkg/kubectl/main/main.go create -f assets/pod_create_test2.yaml

go run pkg/kubectl/main/main.go delete -f assets/pod_create_test1.yaml
```

