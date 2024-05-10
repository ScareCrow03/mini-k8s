#将当前位置设置为工作目录
NEW_WORKDIR=$(pwd)

# 删除旧的WORKDIR环境变量
sed -i '/^export WORKDIR=/d' ~/.bashrc

# 设置新的WORKDIR环境变量
echo "export WORKDIR=$NEW_WORKDIR" >> ~/.bashrc

# 使设置的环境变量立即生效
source ~/.bashrc

# 重启终端
exec bash

