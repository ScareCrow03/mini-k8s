#!/bin/bash
# 用于master节点初始化
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

echo $SCRIPT_DIR
CUR_WORK_DIR=$SCRIPT_DIR/..

CUR_MASTER_IP=192.168.1.5

docker run -d --restart always --name registry -p 5000:5000 -v $CUR_WORK_DIR/assets/registry:/var/lib/registry registry:2

sudo echo '{  "insecure-registries" : ["192.168.1.5:5000"] }' | sudo tee /etc/docker/daemon.json

sudo systemctl daemon-reload
sudo systemctl restart docker

cd $CUR_WORK_DIR/serverlessdockerfile/baseimage
docker build -t $CUR_MASTER_IP:5000/baseserver:latest .
docker push $CUR_MASTER_IP:5000/baseserver:latest

curl -XGET http://$CUR_MASTER_IP:5000/v2/_catalog