#!/bin/bash

# 这个脚本用于安装flannel网络插件，需要合适地配置etcd服务的地址，master节点和worker节点的行为也不同
ETCD_ENDPOINTS = "127.0.0.1:2379"

if systemctl is-active --quiet flanneld; then
    echo "Flannel is already running. No need to install."
else
    echo "Flannel is not running. Starting installation..."
    
    # master节点的etcd服务需要监听所有网卡，否则其他机器的flannel启动后连不上etcd服务
    # 因为etcd被systemd管理成了一个服务，所以需要修改/etc/systemd/system/etcd.service的内容为如下；请注意缩进问题，第二个EOF必须在shell脚本的最左侧
    sudo tee /etc/systemd/system/etcd.service << EOF
    [Unit]
    Description=etcd
    After=network.target

    [Service]
    ExecStart=/usr/local/bin/etcd --listen-client-urls="http://0.0.0.0:2379" --advertise-client-urls="http://0.0.0.0:2379"
    Restart=on-failure
    Type=notify

    [Install]
    WantedBy=default.target
EOF

    # 修改后，从磁盘重新加载systemctl配置，重启etcd服务
    systemctl daemon-reload
    systemctl restart etcd


    etcdctl put /coreos.com/network/config '{"Network": "10.5.0.0/16", "SubnetLen": 24, "SubnetMin": "10.5.1.0","SubnetMax": "10.5.20.0", "Backend": {"Type": "vxlan"}}'

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
fi
systemctl status flanneld