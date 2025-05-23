## cmd

一键删除所有容器

```bash
docker stop $(docker ps -aq) && docker rm $(docker ps -aq)
```

```bash
docker stop [容器名/容器id]

docker kill [容器名/容器id]

./script/clear_etcd.sh

etcdctl get --prefix /registry

// 进入某容器
docker exec -it container_name bin/sh

```


启动

```bash
go run pkg/apiserver/main/main.go

go run pkg/scheduler/main/main.go

go run pkg/kubelet/main/main.go

su root
source /etc/profile
go run pkg/controller/main/main.go
go run pkg/kubeproxy/main/main.go

go run pkg/prometheus/main/main.go

# 请在已经有环境变量的root用的shell下运行以下普通的go，而不要用sudo、防止环境变量与父shell不一致的问题！
go run pkg/serverless/main/main.go
```



```bash
go run pkg/kubectl/main/main.go get pod

go run pkg/kubectl/main/main.go create -f assets/pod_create_test1.yaml
go run pkg/kubectl/main/main.go delete -f assets/pod_create_test1.yaml

go run pkg/kubectl/main/main.go create -f assets/pod_create_test2.yaml
go run pkg/kubectl/main/main.go delete -f assets/pod_create_test2.yaml

# 验收补充用例
go run pkg/kubectl/main/main.go create -f assets/final_test/my_test_pod1.yaml
go run pkg/kubectl/main/main.go delete -f assets/final_test/my_test_pod1.yaml

go run pkg/kubectl/main/main.go create -f assets/final_test/my_test_pod2.yaml
go run pkg/kubectl/main/main.go delete -f assets/final_test/my_test_pod2.yaml

# 创建service，它管理cpu标签为i9-14900k的pods，暴露NodePort为30001
go run pkg/kubectl/main/main.go create -f assets/service_create_test1.yaml
go run pkg/kubectl/main/main.go delete -f assets/service_create_test1.yaml

go run pkg/kubectl/main/main.go get service
# 访问NodePort
curl 10.181.111.128:30001

go run pkg/kubectl/main/main.go create -f assets/replicaset_create_test1.yaml
go run pkg/kubectl/main/main.go delete -f assets/replicaset_create_test1.yaml

go run pkg/kubectl/main/main.go create -f assets/hpa_test_create.yaml

go run pkg/kubectl/main/main.go delete -f assets/hpa_test_create.yaml

# 验收测试Hpa，先创建replica，然后用svc管理它，再建立一个hpa
# 经测试，设置10并发数/s时，为了让hpa中每个pod的cpu占用率不超过0.5，需要5个pods；这里简单访问一下nginx并不太占内存空间，memory占用几乎不变（实际一个容器启动nginx只占用6MiB左右，可以据此设计目标限制的memory值）
go run pkg/kubectl/main/main.go create -f assets/final_test/for_hpa/my_replica.yaml
go run pkg/kubectl/main/main.go delete -f assets/final_test/for_hpa/my_replica.yaml

go run pkg/kubectl/main/main.go create -f assets/final_test/for_hpa/svc_on_my_replica.yaml
go run pkg/kubectl/main/main.go delete -f assets/final_test/for_hpa/svc_on_my_replica.yaml

go run pkg/kubectl/main/main.go create -f assets/final_test/for_hpa/hpa_on_my_replica.yaml
go run pkg/kubectl/main/main.go delete -f assets/final_test/for_hpa/hpa_on_my_replica.yaml

# 创建一个CR对象，类型是PingSource，需要配合PingSourceController实现按Scheduler发消息的功能
# 请结合default/func1（x+y函数）使用
go run pkg/kubectl/main/main.go create -f assets/test_serverless/test_ping_source1.yaml

go run pkg/kubectl/main/main.go delete -f assets/test_serverless/test_ping_source1.yaml

go run pkg/kubectl/main/main.go create -f assets/test_prometheus/test_prometheus_pod1.yaml

go run pkg/kubectl/main/main.go delete -f assets/test_prometheus/test_prometheus_pod1.yaml

go run pkg/kubectl/main/main.go create -f assets/test_serverless/test_func1.yaml

go run pkg/kubectl/main/main.go create -f assets/test_persistent/pv1.yaml
go run pkg/kubectl/main/main.go delete -f assets/test_persistent/pv1.yaml
go run pkg/kubectl/main/main.go create -f assets/test_persistent/pvc1.yaml
go run pkg/kubectl/main/main.go delete -f assets/test_persistent/pvc1.yaml

go run pkg/kubectl/main/main.go get pv
go run pkg/kubectl/main/main.go get pvc

go run pkg/kubectl/main/main.go create -f assets/test_persistent/pod_web.yaml
# 在/srv/mini-k8s/mypv1/default.mypvc1中执行
echo "Hello, this is mini-k8s PVC mypvc1" > index.html
go run pkg/kubectl/main/main.go get pod
curl pod_web_ip

go run pkg/kubectl/main/main.go create -f assets/test_serverless/test_serverless1.yaml

curl -X POST localhost:8050/triggerFunction/default/func1 -H "Content-Type: application/json" -d '{"x": 123, "y": 789}'

# 以下创建一个workflow，首先需要把它依赖的函数建立出来

go run pkg/kubectl/main/main.go create -f assets/test_serverless/test_workflow/workflow_func1.yaml

curl -X POST localhost:8050/triggerFunction/default/fibonaccifunc -H "Content-Type: application/json" -d '{"x": 0, "y": 1, "i": 1}'

go run  pkg/kubectl/main/main.go create -f assets/test_serverless/test_workflow/test_workflow1.yaml

go run pkg/kubectl/main/main.go get workflow

curl -X POST localhost:8050/triggerWorkflow/default/FibonacciWorkflow


# 测试do while逻辑，一定至少会做一次循环体，默认限制i上限为10时，这会得到i==11的值
curl -X POST localhost:8050/triggerWorkflow/default/FibonacciWorkflow -H "Content-Type: application/json" -d '{"x": 34, "y": 55, "i": 10}'

# prometheus监控
# 构建自定义镜像到本地，这里没有push到公有hub；如果需要多机，建议每个节点都运行一遍
sudo script/build_expose_image.sh

# 运行单个有指标暴露的Pod
go run pkg/kubectl/main/main.go create -f assets/test_prometheus/test_prometheus_pod1.yaml
go run pkg/kubectl/main/main.go delete -f assets/test_prometheus/test_prometheus_pod1.yaml

# 运行多个有指标暴露的Pod
go run pkg/kubectl/main/main.go create -f assets/test_prometheus/test_prometheus_replica.yaml
go run pkg/kubectl/main/main.go delete -f assets/test_prometheus/test_prometheus_replica.yaml
go run pkg/kubectl/main/main.go get pod

# 删除prometheus的TSDB关于某个指标的所有数据
# curl -X POST -g 'http://localhost:9090/api/v1/admin/tsdb/delete_series?match[]={__name__="aaa_my_metric"}'
```

# 以下创建一个复杂的serverless工作流
go run pkg/kubectl/main/main.go create -f assets/test_complex_serverless/extractimgemetadata.yaml

curl -X POST http://localhost:8050/triggerFunction/default/extractimagemetadata -H "Content-Type: application/json" -d '{
  "COUCHDB_URL": "http://admin:123@192.168.183.128:5984",
  "COUCHDB_DBNAME": "image",
  "IMAGE_NAME": "image.jpg",
  "IMAGE_DOCID": "image1"
}'

go run pkg/kubectl/main/main.go create -f assets/test_complex_serverless/handler.yaml

curl -X POST localhost:8050/triggerFunction/default/handler -H "Content-Type: application/json" -d '{
  "COUCHDB_URL": "http://admin:123@192.168.183.128:5984",
  "COUCHDB_DBNAME": "image",
  "IMAGE_NAME": "image.jpg",
  "IMAGE_DOCID": "image1",
  "COUCHDB_LOGDB": "logdb"
}'

go run pkg/kubectl/main/main.go create -f assets/test_complex_serverless/thumbnail.yaml

curl -X POST http://localhost:8050/triggerFunction/default/thumbnail -H "Content-Type: application/json" -d '{
  "COUCHDB_URL": "http://admin:123@10.181.111.128:5984",
  "COUCHDB_DBNAME": "image",
  "IMAGE_NAME": "image.jpg",
  "IMAGE_DOCID": "image1",
  "COUCHDB_LOGDB": "logdb"
}'

go run  pkg/kubectl/main/main.go create -f assets/test_complex_serverless/complex_workflow.yaml

curl -X POST localhost:8050/triggerWorkflow/default/ImageProcessWorkflow

```bash
sudo rabbitmqctl add_user visitor 123456
sudo rabbitmqctl  set_user_tags  visitor  administrator
sudo systemctl start rabbitmq-server
sudo rabbitmqctl set_permissions -p "/" visitor ".*" ".*" ".*"
```