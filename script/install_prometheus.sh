#!/bin/bash

# 这个脚本用于在master节点安装prometheus系统服务
# 请保持主机的9090端口未被占用，否则Prometheus无法启动！
# 2024-5 最新Prometheus版本
PROMETHEUS_VERSION="2.52.0"

# Prometheus配置文件路径
PROMETHEUS_CONFIG_PATH="/etc/prometheus/prometheus.yml"

if systemctl is-active --quiet prometheus; then
    echo "Prometheus is already running. No need to install."
else
    echo "Prometheus is not running. Starting installation..."

    # 下载Prometheus二进制文件，有100M左右
    wget https://github.com/prometheus/prometheus/releases/download/v${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION}.linux-amd64.tar.gz
    # 解压，注意不会自动删除，防止出错时没有文件！
    tar -xzvf prometheus-${PROMETHEUS_VERSION}.linux-amd64.tar.gz
    cd prometheus-${PROMETHEUS_VERSION}.linux-amd64

    # 复制相关文件到指定位置
    sudo cp prometheus /usr/local/bin/
    sudo cp promtool /usr/local/bin/

    sudo mkdir -p /usr/share/prometheus | true
    sudo cp ./consoles /usr/share/prometheus/consoles -r
    sudo cp ./console_libraries /usr/share/prometheus/console_libraries -r

    # 创建Prometheus配置文件目录
    sudo mkdir -p $(dirname ${PROMETHEUS_CONFIG_PATH})

    # 复制Prometheus配置文件到指定位置
    sudo cp prometheus.yml ${PROMETHEUS_CONFIG_PATH}
    # 修改配置文件权限，否则无法正确写入
    sudo chmod 777 ${PROMETHEUS_CONFIG_PATH}
   
    # 创建Prometheus的systemd服务文件，指定允许热更新，允许admin-api（比如删掉时序数据库中的一些序列）
    sudo tee /etc/systemd/system/prometheus.service << EOF
[Unit]
Description=Prometheus
After=network.target

[Service]
ExecStart=/usr/local/bin/prometheus \
  --config.file=${PROMETHEUS_CONFIG_PATH} \
  --storage.tsdb.path=/var/lib/prometheus/ \
  --web.console.libraries=/usr/share/prometheus/console_libraries \
  --web.console.templates=/usr/share/prometheus/consoles \
  --web.enable-lifecycle \
  --web.enable-admin-api
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

    # 重新加载systemd，启动Prometheus服务并设置为开机启动
    sudo systemctl daemon-reload
    sudo systemctl start prometheus
    sudo systemctl enable prometheus
fi

systemctl status prometheus
