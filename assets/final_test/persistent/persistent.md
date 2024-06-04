## 持久化存储

### 12
手动和自动创建PV、展示PVC配置文件

```bash
go run pkg/apiserver/main/main.go
go run pkg/scheduler/main/main.go
go run pkg/kubelet/main/main.go

go run pkg/kubectl/main/main.go create -f assets/test_persistent/pv1.yaml
go run pkg/kubectl/main/main.go delete -f assets/test_persistent/pv1.yaml

go run pkg/kubectl/main/main.go create -f assets/test_persistent/pvc1.yaml
go run pkg/kubectl/main/main.go delete -f assets/test_persistent/pvc1.yaml

go run pkg/kubectl/main/main.go create -f assets/test_persistent/pvc2.yaml
go run pkg/kubectl/main/main.go delete -f assets/test_persistent/pvc2.yaml

go run pkg/kubectl/main/main.go create -f assets/final_test/persistent/pod_web1.yaml

cd /srv/mini-k8s/mypv1
# 在/srv/mini-k8s/mypv1/default.mypvc1中执行
echo "Hello, this is mini-k8s PVC mypvc1." > index.html
echo "Hello again!" >> index.html

go run pkg/kubectl/main/main.go get pod
curl pod_web_ip
```

### 3
展⽰Pod和PV解绑之后重新绑定
先删除之前创建的Pod，再重新创建

```bash
go run pkg/kubectl/main/main.go delete -f assets/final_test/persistent/pod_web1.yaml

go run pkg/kubectl/main/main.go get pod

go run pkg/kubectl/main/main.go create -f assets/final_test/persistent/pod_web1.yaml
```

### 4
多机持久化存储
启动另一个节点，再启动四个Pod，均使用同一个PVC

```bash
go run pkg/kubelet/main/main.go

go run pkg/kubectl/main/main.go create -f assets/final_test/persistent/pod_web1.yaml

go run pkg/kubectl/main/main.go create -f assets/final_test/persistent/pod_web2.yaml

go run pkg/kubectl/main/main.go create -f assets/final_test/persistent/pod_web3.yaml

go run pkg/kubectl/main/main.go create -f assets/final_test/persistent/pod_web4.yaml

go run pkg/kubectl/main/main.go create -f assets/final_test/persistent/pod_web5.yaml

go run pkg/kubectl/main/main.go get pod

curl pod_ip
```