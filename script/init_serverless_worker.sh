#!/bin/bash
# 用于worker节点初始化
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

echo $SCRIPT_DIR
CUR_WORK_DIR=$SCRIPT_DIR/..

CUR_MASTER_IP=192.168.1.5

sudo echo '{  "insecure-registries" : ["192.168.1.5:5000"] }' | sudo tee /etc/docker/daemon.json

sudo systemctl daemon-reload
sudo systemctl restart docker

curl -XGET http://$CUR_MASTER_IP:5000/v2/_catalog

docker pull $CUR_MASTER_IP:5000/baseserver:latest

