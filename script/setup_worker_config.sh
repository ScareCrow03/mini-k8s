#将当前位置设置为工作目录
NEW_WORKDIR=$(pwd)

# 删除旧的WORKDIR环境变量
sed -i '/^export WORKDIR=/d' ~/.bashrc

# 设置新的WORKDIR环境变量
echo "export WORKDIR=$NEW_WORKDIR" >> ~/.bashrc

# 使设置的环境变量立即生效
source ~/.bashrc


# 检查环境变量是否设置了主机IP地址，如果未设置则使用默认值
if [ -z "$MASTER_IP" ]; then
    default_master_ip=""  # 默认主机IP地址为空
else
    default_master_ip="$MASTER_IP"
fi

# 提示用户输入主机IP地址，如果用户未输入则使用默认值
read -p "请输入MASTER节点IP地址 [默认为${default_master_ip:-暂无默认值}]: " master_ip

# 如果用户没有输入任何内容，则使用默认值
if [ -z "$master_ip" ]; then
    master_ip="$default_master_ip"
fi

# 如果用户输入了非空格的内容，将其设置为环境变量
if [ -n "$master_ip" ] && [ "$master_ip" != " " ]; then
    sed -i '/^export MASTER_IP=/d' ~/.bashrc
    echo "export MASTER_IP=$master_ip" >> ~/.bashrc
fi

read -p "请输入本node名称: " node_name
if [ -z "$node_name" ]; then
    echo "node名称不能为空"
    exit 1
fi

#写入yaml文件,如果存在文件先删除
rm -f $NEW_WORKDIR/assets/worker-config.yaml
echo "apiServerAddress: http://$master_ip:8080" > $NEW_WORKDIR/assets/worker-config.yaml
echo "name: $node_name" >> $NEW_WORKDIR/assets/worker-config.yaml
echo "roles: worker" >> $NEW_WORKDIR/assets/worker-config.yaml
echo "version: 1.0" >> $NEW_WORKDIR/assets/worker-config.yaml
echo "成功生成配置文件，如需修改，在$NEW_WORKDIR/assets/worker-config.yaml"
source ~/.bashrc

# 重启终端
exec bash