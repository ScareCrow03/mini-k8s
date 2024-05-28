package net_util

import (
	"fmt"
	"net"
	"strings"
)

// 获取本机的物理网卡IP，区别方式是物理网卡可以连接到外部网络，故建立连接后反过来获取本地IP即可
func GetNodeIP() (string, error) {
	// 创建一个到外部网络的连接
	conn, err := net.Dial("udp", "www.baidu.com:80")
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer conn.Close()

	// 获取连接的本地地址
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// 打印本地IP地址
	fmt.Println(localAddr.IP)

	return strings.TrimSpace(localAddr.IP.String()), nil
}
