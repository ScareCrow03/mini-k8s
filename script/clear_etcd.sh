#!/bin/bash

# if [ $# -eq 0 ]; then
#   echo "Usage: $0 <prefix>"
#   exit 1
# fi

substring="registry"

echo "Deleting etcd storage with prefix: $substring"

# 获取 etcd 中所有键值
keys=$(etcdctl get --prefix /registry)

# 循环删除包含指定字符串的键值
for key in $keys; do
  if [[ $key == *$substring* ]]; then
    echo  "Delete key: $key"
    etcdctl del $key
  fi
done

echo "Done!"

# 只删除名称中包含minik8s字符串的容器，不删除其他容器！
docker stop $(docker ps -a --filter "name=minik8s" -q) && docker rm $(docker ps -a --filter "name=mini-k8s" -q)
