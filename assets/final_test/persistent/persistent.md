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

go run pkg/kubectl/main/main.go create -f assets/test_persistent/pod_web.yaml
# 在/srv/mini-k8s/mypv1/default.mypvc1中执行
echo "Hello, this is mini-k8s PVC mypvc1" > index.html
go run pkg/kubectl/main/main.go get pod
curl pod_web_ip
```

### 3
展⽰Pod和PV解绑之后重新绑定

```bash

```