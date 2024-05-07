#!/bin/bash

# 这个脚本记录一些杂项，在安装完成所有依赖后执行它

# 开启Linux的IP转发功能。当IP转发被开启时，Linux系统可以将收到的数据包转发给其他网络设备，从而充当路由器的角色
sudo sysctl --write net.ipv4.ip_forward=1
# 启动内核模块br_netfilter，这个模块允许iptables的规则应用到桥接的数据包上，从而实现对桥接网络的过滤和控制。
sudo modprobe br_netfilter

# 让桥接设备在进行二层转发时也去调用iptables配置的三层规则。这样可以解决在同一节点上，一个Pod去访问不包含该Pod的Service的问题。
sudo sysctl --write net.bridge.bridge-nf-call-iptables=1

# 开启mini-cni0网桥（这是flannel注册到docker网桥上后多出来的主机设备）的混杂模式；普通模式下，网卡只接收发给本机的包（包括广播包）传递给上层程序，其它的包一律丢弃。混杂模式下，网卡会接收所有经过的数据包，包括那些不是发给本机的包，即不验证MAC地址。
sudo ip link set mini-cni0 promisc on

# 开启了IP虚拟服务器（IPVS）的连接跟踪功能。连接跟踪可以记录和维护每个连接的状态信息，从而实现更复杂的网络功能，如NAT、防火墙等
sudo sysctl --write net.ipv4.vs.conntrack=1

