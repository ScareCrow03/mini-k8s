#! /bin/bash
## 一些预处理操作

# 更新ubuntu的包管理器
sudo apt-get update

# 安装wget用于从url下载文件
sudo apt-get install wget

## Go依赖安装
# 下载最新版本1.22.2的go语言包，于2024-4-4在阿里云上更新
GO_SRC_URL="https://mirrors.aliyun.com/golang/go1.22.2.linux-amd64.tar.gz"

# 检查Go是否已经安装
if command -v go >/dev/null 2>&1; then
    echo "Go is already installed. Skipping installation."
else
    echo "Go is not installed. Starting installation."

    # 下载Go的tar包到当前的目录
    wget -q "$GO_SRC_URL" -O go.tar.gz

    # 解压tar包到指定的安装目录；则此时go的二进制文件路径为/usr/local/go/bin
    sudo tar -C "/usr/local" -xzf go.tar.gz

    # 删除下载的tar包
    rm go.tar.gz

    # 将配置PATH的脚本go_path.sh放在profile.d下，这样在所有用户的shell加载时会正确执行这些脚本，并将路径添加到PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go_path.sh > /dev/null
    sudo chmod 777 /etc/profile.d/go_path.sh

    # 在当前shell中加载上述脚本
    source /etc/profile.d/go_path.sh
    
    echo "Go has been installed and configured."
fi
# 显示版本
go version

## docker依赖安装
# 检查Docker是否已经安装
if command -v docker >/dev/null 2>&1; then
    echo "Docker is already installed. Skipping installation."
else
    echo "Docker is not installed. Starting installation."

    # 更新包列表
    sudo apt-get update

    # 安装Docker的依赖
    sudo apt-get install \
            apt-transport-https \
            ca-certificates \
            curl \
            gnupg \
            lsb-release

    # 添加阿里云的GPG密钥
    curl -fsSL http://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | sudo apt-key add -


    # 设置docker稳定的存储库为阿里云
    sudo add-apt-repository "deb [arch=amd64] http://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable"

    ## 这是官方存储库，但是jcloud经常连不上它
    # sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" 

    # 更新包列表
    sudo apt-get update

    # 查看当前包列表中有哪些docker-ce版本；如果上述配置正常，这一步查询结果应该不为空！
    apt list -a docker-ce

    # 安装特定版本的Docker引擎，这个版本号应该在上述的查询中出现过
    sudo apt-get install docker-ce=5:24.0.9-1~ubuntu.20.04~focal docker-ce-cli=5:24.0.9-1~ubuntu.20.04~focal containerd.io
    
    # 让Docker在启动时自动运行
    sudo systemctl enable docker

    # 将用户添加到docker组，避免每次运行Docker命令时都需要输入sudo
    sudo usermod -aG docker $USER || true

    # 刷新组成员资格
    newgrp docker || true

    # 查看docker系统服务状态，应该是active(running)
    sudo systemctl status docker

    echo "Docker has been installed and configured."
fi

docker -v
echo "Docker is ready to use without sudo."


## etcd依赖安装
# 宿主机端口2379
# 安装etcd v3.5.13，于2024-3更新；安装目录在usr/local/bin

# etcd的版本
ETCD_VER=v3.5.13

# 下载etcd的url源，默认为GITHUB
GOOGLE_URL=https://storage.googleapis.com/etcd
GITHUB_URL=https://github.com/etcd-io/etcd/releases/download
DOWNLOAD_URL=${GITHUB_URL}

# 检查etcd是否已经安装
if command -v etcd >/dev/null 2>&1; then

    echo "etcd is already installed. Skipping installation."
else
    echo "etcd is not installed. Starting installation."

    rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
    rm -rf /tmp/test-etcd && mkdir -p /tmp/test-etcd

    curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
    tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/test-etcd --strip-components=1

    # 移动etcd和etcdctl到/usr/local/bin目录
    mv /tmp/test-etcd/etcd /usr/local/bin
    mv /tmp/test-etcd/etcdctl /usr/local/bin

    echo "etcd has been installed and configured."

    # 创建etcd的systemd服务文件，用于开机启动
    sudo tee /etc/systemd/system/etcd.service > /dev/null <<EOF
    [Unit]
    Description=etcd service
    After=network.target

    [Service]
    ExecStart=/usr/local/bin/etcd
    Restart=on-failure
    Type=notify

    [Install]
    WantedBy=default.target
EOF

    # 重新加载systemd的配置
    sudo systemctl daemon-reload

    # 启动etcd服务
    sudo systemctl start etcd

    # 设置etcd服务在启动时自动运行
    sudo systemctl enable etcd

    echo "etcd service has been configured and started."
fi
# 显示版本
etcd --version

### rabbitMQ安装
# 检查 RabbitMQ 服务是否在运行
if systemctl is-active --quiet rabbitmq-server; then
    echo "RabbitMQ is already running. No need to install."
else
    echo "RabbitMQ is not running. Starting installation..."
    sudo apt update
    sudo apt install rabbitmq-server
    sudo systemctl start rabbitmq-server
    sudo systemctl enable rabbitmq-server
    echo "RabbitMQ installation completed and the service is now running."
fi

### 网络依赖安装
sudo apt-get install iptables
sudo apt-get install ipset
sudo apt-get install ipvsadm