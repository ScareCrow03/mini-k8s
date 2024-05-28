#!/bin/bash

# 这个脚本用于修复之前注册flannel到docker网桥时，未指定对应的主机网络设备的问题，其作用是删除旧的，再新建一份


container_ids=$(docker network inspect -f '{{range $k, $v := .Containers}}{{$v.Name}} {{end}}' flannel)

# 对于每个容器ID，停止并删除对应的容器
for id in $container_ids
do
    docker rm -f $id
done

# 删除旧flannel网络
docker network rm flannel

sudo /usr/local/bin/mk-docker-opts.sh # 在flannel正常运行时，再运行这个tar包自带的脚本，它会提取flannel的实时状态，并生成下面的、用于配置docker的环境变量文件

source /run/flannel/subnet.env # 加载这个环境变量文件到当前shell
docker network create --attachable=true --subnet=${FLANNEL_SUBNET} -o "com.docker.network.driver.mtu"=${FLANNEL_MTU} -o "com.docker.network.bridge.name"="mini-cni0" flannel # 为docker创建一个flannel网络，它使用了这个flannel分配的子网；让这个网络在主机上的设备名称为mini-cni0

# 查看本节点持有的flannel网段，默认一个节点上可分配256个IP；这是根据上述etcd写入的配置自动生成的、在10.5.0.0/16的IP全空间下的一小块；一个Node的运行时信息、可以包含由flannel管理的这个网段
docker network inspect flannel | grep Subnet | awk -F '\"' '{print $4}'