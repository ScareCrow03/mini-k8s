#!/bin/bash

# 这个脚本用于安装flannel网络插件，需要合适地配置etcd服务的地址，master节点和worker节点的行为也不同；worker节点一般本身没有etcd服务，需要连接master节点上的etcd服务！
ETCD_ENDPOINTS = "127.0.0.1:2379"

if systemctl is-active --quiet flanneld; then
    echo "Flannel is already running. No need to install."
else
    echo "Flannel is not running. Starting installation..."
    
    # 下载flannel安装包，解压并复制到 /usr/local/bin/目录下（这个目录已经在PATH里，方便在任何地方启动可执行文件），添加脚本执行权限
    wget https://github.com/flannel-io/flannel/releases/download/v0.25.1/flannel-v0.25.1-linux-amd64.tar.gz
    mkdir ./flannel_install
    tar -xzvf flannel-v0.25.1-linux-amd64.tar.gz -C ./flannel_install
    cd ./flannel_install

    sudo cp ./flanneld /usr/local/bin/
    sudo cp ./mk-docker-opts.sh /usr/local/bin/
    sudo chmod 777 /usr/local/bin/mk-docker-opts.sh

    # 创建flanneld的systemd服务文件，用于开机启动
    sudo tee /etc/systemd/system/flanneld.service << EOF
    [Unit]
    Description=Flannel
    After=network.target
    After=network-online.target
    Wants=network-online.target

    [Service]
    # 设置服务启动的命令，注意应该与具体的etcd地址保持一致！
    ExecStart=/usr/local/bin/flanneld --ip-masq --kube-subnet-mgr=false --etcd-endpoints=${ETCD_ENDPOINTS}
    Restart=on-failure

    [Install]
    WantedBy=multi-user.target
EOF

    # 创建这个服务单元文件后，重新加载，启动flanneld服务并设置为开机启动
    sudo systemctl daemon-reload
    sudo systemctl start flanneld
    sudo systemctl enable flanneld

    sudo /usr/local/bin/mk-docker-opts.sh # 在flannel正常运行时，再运行这个tar包自带的脚本，它会提取flannel的实时状态，并生成下面的、用于配置docker的环境变量文件

    source /run/flannel/subnet.env # 加载这个环境变量文件到当前shell
    docker network create --attachable=true --subnet=${FLANNEL_SUBNET} -o "com.docker.network.driver.mtu"=${FLANNEL_MTU} flannel # 为docker创建一个flannel网络，它使用了这个flannel分配的子网

    # 查看本节点持有的flannel网段，默认一个节点上可分配256个IP；
    docker network inspect flannel | grep Subnet | awk -F '\"' '{print $4}'
EOF
