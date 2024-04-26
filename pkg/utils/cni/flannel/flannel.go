package flannelClient

import (
	"fmt"
	"os/exec"
	"strings"
)

// 给定一个容器ID，查询其在cni网络中的IP地址；我们假定一个容器只加入了一个网络
func LookupIP(containerID string) (string, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerID)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
	}
	return string(strings.TrimSpace(string(output))), err
}
