### 安装配置flannel

用于单节点上运行etcd服务，多节点搭建flannel网络的配置

```shell
# 请按安装依赖脚本，前置安装etcd，使之运行在2379端口

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

# 在已经安装好etcd的主机上写入一个串，指定关于多机flannel网络分配ip的一些配置
etcdctl put /coreos.com/network/config '{"Network": "10.5.0.0/16", "SubnetLen": 24, "SubnetMin": "10.5.1.0","SubnetMax": "10.5.20.0", "Backend": {"Type": "vxlan"}}'

# "Network"参数，指定flannel网络的IP全空间是"10.5.0.0/16"，共2^(32-16)=66536个IP；
# "SubnetLen"参数，指定每个节点的flannel子网长度为mask=24的子网，共2^(32-24)=256个IP
# "SubnetMin", "SubnetMax"指定flannel为节点分配子网时可以使用的IP地址范围，可见这里设置为至多支持19个节点共用一个flannel网络
# Backend": {"Type": "vxlan"}参数，指定flannel后端使用vxlan

# 这个串一般不会发生变化，如果发生变化，那么需要手动重启各flannel进程、删除原来的docker网络flannel再建立一个新的

# 下载flannel安装包，解压并复制到 /usr/local/bin/目录下（这个目录已经在PATH里，方便在任何地方启动可执行文件），添加脚本执行权限
wget https://github.com/flannel-io/flannel/releases/download/v0.25.1/flannel-v0.25.1-linux-amd64.tar.gz
mkdir ./flannel_install
tar -xzvf flannel-v0.25.1-linux-amd64.tar.gz -C ./flannel_install
cd ./flannel_install

sudo cp ./flanneld /usr/local/bin/
sudo cp ./mk-docker-opts.sh /usr/local/bin/
sudo chmod 777 /usr/local/bin/mk-docker-opts.sh

# 以下为运行flannel进程的两种方式，只需要选取其中一种
# 前台运行会阻塞这个shell；注意--etcd-endpoints需要设置为真实的etcd服务地址，尤其是多节点、单etcd入口的情况；建议选这种来做调试
sudo flanneld  --ip-masq --kube-subnet-mgr=false --etcd-endpoints=http://127.0.0.1:2379 

# 后台运行、并将输出重定向到nohup.out，查看后台进程的日志可以通过cat nohup.out
sudo nohup flanneld  --ip-masq --kube-subnet-mgr=false --etcd-endpoints=http://127.0.0.1:2379 &

jobs # 查看这个进程是否在当前shell启动了
# 按ctrl-c发送SIGINT信号，可以杀掉这些后台进程；


# 以下配置docker网络
# docker network rm flannel # 第一次配置不需要删除flannel网络，如果是重新配置，手动删一下旧的flannel；此时需要没有旧的容器连上这个flannel网络，所以改动一次很麻烦，基本不改动
sudo /usr/local/bin/mk-docker-opts.sh # 在flannel正常运行时，再运行这个tar包自带的脚本，它会提取flannel的实时状态，并生成下面的、用于配置docker的环境变量文件

source /run/flannel/subnet.env # 加载这个环境变量文件到当前shell
docker network create --attachable=true --subnet=${FLANNEL_SUBNET} -o "com.docker.network.driver.mtu"=${FLANNEL_MTU} flannel # 为docker创建一个flannel网络，它使用了这个flannel分配的子网

# 现在docker network ls可以看到flannel网络
# 可以尝试创建一个容器，指定加入该flannel网络的管理
# docker run -it --net=flannel --name=test_curl -d curlimages/curl sh
# 注意，上述flannel网络在第一次创建后就会在docker network列表中一直存在；但是为了让它可用，还是需要启动flanneld进程
# 现在，进入容器后使用ifconfig查看，可以看到eth0网卡获得了flanneld分配的IP，而不是以172开头的、docker0网桥的默认分配IP

# 查看本节点持有的flannel网段，默认一个节点上可分配256个IP；这是根据上述etcd写入的配置自动生成的、在10.5.0.0/16的IP全空间下的一小块；一个Node的运行时信息、可以包含由flannel管理的这个网段
docker network inspect flannel | grep Subnet | awk -F '\"' '{print $4}'
```



在以上调试正确后，可以将flannel设置为一个服务，使之开机启动；否则，每次需要flannel网络时，必须新开一个shell来运行它，还不能退出

```shell
sudo tee /etc/systemd/system/flanneld.service << EOF
[Unit]
Description=Flannel
After=network.target
After=network-online.target
Wants=network-online.target

[Service]
# 设置服务启动的命令，注意应该与具体的etcd地址保持一致！
ExecStart=/usr/local/bin/flanneld --ip-masq --kube-subnet-mgr=false --etcd-endpoints=http://127.0.0.1:2379
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# 创建这个服务单元文件后，重新加载，启动flanneld服务并设置为开机启动
sudo systemctl daemon-reload
sudo systemctl start flanneld
sudo systemctl enable flanneld

# 查看最近的100条日志输出，新的日志在上方；如果希望旧的日志在上方，可以去掉-r参数
sudo journalctl -u flanneld -n 100 -r
```






