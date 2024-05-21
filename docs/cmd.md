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

go run pkg/controller/main/main.go

sudo go run pkg/kubeproxy/main/main.go

go run pkg/prometheus/main/main.go
```



```bash
go run pkg/kubectl/main/main.go get pod

go run pkg/kubectl/main/main.go create -f assets/pod_create_test1.yaml
go run pkg/kubectl/main/main.go delete -f assets/pod_create_test1.yaml

go run pkg/kubectl/main/main.go create -f assets/pod_create_test2.yaml
go run pkg/kubectl/main/main.go delete -f assets/pod_create_test2.yaml

go run pkg/kubectl/main/main.go create -f assets/service_create_test1.yaml
go run pkg/kubectl/main/main.go delete -f assets/service_create_test1.yaml

go run pkg/kubectl/main/main.go create -f assets/replicaset_create_test1.yaml
go run pkg/kubectl/main/main.go delete -f assets/replicaset_create_test1.yaml

go run pkg/kubectl/main/main.go create -f assets/hpa_test_create.yaml

go run pkg/kubectl/main/main.go delete -f assets/hpa_test_create.yaml

# 创建一个CR对象，类型是PingSource，需要配合PingSourceController实现按Scheduler发消息的功能
go run pkg/kubectl/main/main.go create -f assets/test_serverless/test_ping_source1.yaml

go run pkg/kubectl/main/main.go delete -f assets/test_serverless/test_ping_source1.yaml

go run pkg/kubectl/main/main.go create -f assets/test_prometheus/test_prometheus_pod1.yaml

go run pkg/kubectl/main/main.go delete -f assets/test_prometheus/test_prometheus_pod1.yaml
```



```bash
sudo rabbitmqctl add_user visitor 123456
sudo rabbitmqctl  set_user_tags  visitor  administrator
sudo systemctl start rabbitmq-server
sudo rabbitmqctl set_permissions -p "/" visitor ".*" ".*" ".*"
```