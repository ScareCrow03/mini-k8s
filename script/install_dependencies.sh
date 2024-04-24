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

    # 添加Docker的官方GPG密钥
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

    # 设置稳定的存储库
    echo \
      "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    # 更新包列表
    sudo apt-get update

    # 安装Docker引擎
    sudo apt-get install docker-ce docker-ce-cli containerd.io

    echo "Docker has been installed and configured."
fi
# 显示版本
docker -v

# 以下操作即使docker已经存在于本地也会做一遍；因为每次登录的user可能不同
# 让Docker在启动时自动运行
sudo systemctl enable docker

# 将用户添加到docker组，避免每次运行Docker命令时都需要输入sudo
sudo usermod -aG docker $USER

# 刷新组成员资格
newgrp docker

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


### weave依赖安装
if command -v weave >/dev/null 2>&1; then
    echo "Weave is already installed. Skipping installation."
else
    echo "Weave is not installed. Starting installation."

    # 安装Weave，它是以容器的方式运行的，所以必须先安装docker
    sudo wget -O /usr/local/bin/weave https://raw.githubusercontent.com/zettio/weave/master/weave
    sudo chmod a+x /usr/local/bin/weave

    weave launch 
    echo "Weave has been installed and configured."
fi