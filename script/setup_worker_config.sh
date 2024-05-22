#将当前位置设置为工作目录
NEW_WORKDIR=$(pwd)

# 删除旧的WORKDIR环境变量
sed -i '/^export WORKDIR=/d' /etc/profile

# 设置新的WORKDIR环境变量
echo "export WORKDIR=$NEW_WORKDIR" >> /etc/profile

# 使设置的环境变量立即生效
source /etc/profile


# 检查环境变量是否设置了主机IP地址，如果未设置则使用默认值
if [ -z "$MASTER_IP" ]; then
    default_master_ip=""  # 默认主机IP地址为空
else
    default_master_ip="$MASTER_IP"
fi

# 检查环境变量是否设置了节点名字，如果未设置则使用默认值
if [ -z "$NODENAME" ]; then
    default_node_name=""  # 默认主机IP地址为空
else
    default_node_name="$NODENAME"
fi

# 提示用户输入主机IP地址，如果用户未输入则使用默认值
read -p "请输入MASTER节点IP地址 [默认为${default_master_ip:-暂无默认值}]: " master_ip

# 如果用户没有输入任何内容，则使用默认值
if [ -z "$master_ip" ]; then
    master_ip="$default_master_ip"
fi

# 如果用户输入了非空格的内容，将其设置为环境变量
if [ -n "$master_ip" ] && [ "$master_ip" != " " ]; then
    sed -i '/^export MASTER_IP=/d' /etc/profile
    echo "export MASTER_IP=$master_ip" >> /etc/profile
fi

read -p "请输入本node名称 [默认为${default_node_name:-暂无默认值}]: " node_name

if [ -z "$node_name" ]; then
    node_name="$default_node_name"
fi

# 如果用户输入了非空格的内容，将其设置为环境变量
if [ -n "$node_name" ] && [ "$node_name" != " " ]; then
    sed -i '/^export NODENAME=/d' /etc/profile
    echo "export NODENAME=$node_name" >> /etc/profile
fi

#写入yaml文件,如果存在文件先删除
rm -f $NEW_WORKDIR/assets/worker-config.yaml
echo "apiServerAddress: http://$master_ip:8080" > $NEW_WORKDIR/assets/worker-config.yaml
echo "name: $node_name" >> $NEW_WORKDIR/assets/worker-config.yaml
echo "roles: worker" >> $NEW_WORKDIR/assets/worker-config.yaml
echo "version: 1.0" >> $NEW_WORKDIR/assets/worker-config.yaml
echo "成功生成配置文件，如需修改，在$NEW_WORKDIR/assets/worker-config.yaml"
source /etc/profile


#go.sh#
# 将配置PATH的脚本go_path.sh放在profile.d下，这样在所有用户的shell加载时会正确执行这些脚本，并将路径添加到PATH
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go_path.sh > /dev/null
sudo chmod 777 /etc/profile.d/go_path.sh
# 在当前shell中加载上述脚本
source /etc/profile.d/go_path.sh

###extra_setting.sh

# 开启Linux的IP转发功能。当IP转发被开启时，Linux系统可以将收到的数据包转发给其他网络设备，从而充当路由器的角色
sudo sysctl --write net.ipv4.ip_forward=1
# 启动内核模块br_netfilter，这个模块允许iptables的规则应用到桥接的数据包上，从而实现对桥接网络的过滤和控制。
sudo modprobe br_netfilter
# 以及启动其他一些内核模块
sudo modprobe -- ip_vs
sudo modprobe -- ip_vs_rr
sudo modprobe -- ip_vs_wrr
sudo modprobe -- ip_vs_sh
sudo modprobe -- nf_conntrack


# 让桥接设备在进行二层转发时也去调用iptables配置的三层规则。这样可以解决在同一节点上，一个Pod去访问不包含该Pod的Service的问题。
sudo sysctl --write net.bridge.bridge-nf-call-iptables=1

# 开启mini-cni0网桥（这是flannel注册到docker网桥上后多出来的主机设备）的混杂模式；普通模式下，网卡只接收发给本机的包（包括广播包）传递给上层程序，其它的包一律丢弃。混杂模式下，网卡会接收所有经过的数据包，包括那些不是发给本机的包，即不验证MAC地址。
sudo ip link set mini-cni0 promisc on

# 开启了IP虚拟服务器（IPVS）的连接跟踪功能。连接跟踪可以记录和维护每个连接的状态信息，从而实现更复杂的网络功能，如NAT、防火墙等
sudo sysctl --write net.ipv4.vs.conntrack=1
# 重启终端
exec bash