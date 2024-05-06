package flannelClient

import (
	"bufio"
	"fmt"
	"os"
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

// 查看本Node上，flannel网络为其分配的PodIP范围
func GetPodIpRangeOnNode() (string, error) {
	// 这个文件在配置好flannel环境后会自动生成，后续以它为准读取本Node的PodIP范围
	file, err := os.Open("/run/flannel/subnet.env")
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 如果某行以FLANNEL_SUBNET=开头，那么就返回该行去掉FLANNEL_SUBNET=后的部分。如果在文件中没有找到FLANNEL_SUBNET=，那么函数会返回一个空字符串。
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "FLANNEL_SUBNET=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "FLANNEL_SUBNET=")), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}
