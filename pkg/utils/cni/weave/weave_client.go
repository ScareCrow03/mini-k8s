package weaveClient

import (
	"mini-k8s/pkg/logger"
	"os/exec"
	"strings"
)

// 这个包的相关测试放在runtime测试中，因为防止循环引用
// 以下操作为直接在命令行中执行，而不是使用SDK
// 给定一个容器ID，将其附加到Weave网络上，并返回分配的IP地址
// 并不需要长连接，故不做其他包装
func AttachCtr(containerID string) (string, error) {
	output, err := exec.Command("weave", "attach", containerID).Output()
	if err != nil {
		logger.KError("Failed to attach container %s to weave network: %v", containerID, err)
		return "", err
	}
	// 输出时请把空格去掉！
	return strings.TrimSpace(string(output)), nil
}

// 给定一个容器ID，查询其在weave网络中的IP地址
func LookupIP(containerID string) (string, error) {
	output, err := exec.Command("weave", "ps", containerID).Output()
	if err != nil {
		logger.KError("Failed to attach container %s to weave network: %v", containerID, err)
		return "", err
	}
	// 会返回3个东西，第一个是containerID（等于参数），第二个是MAC形如7e:05:0d:ce:3a:5f，第三个是IP/mask形如10.32.0.1/12；彼此通过空格分隔
	parts := strings.Split(string(output), " ")
	if len(parts) < 3 {
		logger.KWarning("Unexpected weave ps output: %s, maybe no such container", string(output))
		return "", nil
	}
	ipAndMask := parts[2]
	ip := strings.Split(ipAndMask, "/")[0]

	return strings.TrimSpace(ip), nil
}

// 给定一个容器ID，将其从Weave网络上分离，回收IP地址
func DetachCtr(containerID string) error {
	_, err := exec.Command("weave", "detach", containerID).Output()
	if err != nil {
		logger.KError("Failed to detach container %s to weave network: %v", containerID, err)
	}
	return err
}

// TODO: 不同node之间使用weave互联时，是否需要实现其他的方法例如connect、forget的包装？
