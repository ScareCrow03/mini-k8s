#!/bin/bash

# 用apt-get安装grafana为一个系统服务
# 如果只想简单使用，可以考虑直接以容器形式运行grafana

# 导入官方GPG key
sudo mkdir -p /etc/apt/keyrings/
wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null

# 设置grafana稳定版本的源
echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee -a /etc/apt/sources.list.d/grafana.list

# Updates the list of available packages
sudo apt-get update

sudo apt-get install grafana

# 将grafana部署为系统服务，允许开机启动
sudo /bin/systemctl daemon-reload
sudo /bin/systemctl enable grafana-server
sudo /bin/systemctl start grafana-server

## 在本机shell运行，重置grafana管理员密码为给定串
# grafana-cli admin reset-admin-password ${your_password}