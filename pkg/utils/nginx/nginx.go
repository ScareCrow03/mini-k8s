package nginx

import (
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/protocol"
	"os"
	"strconv"
)

func WriteNginxConf(dns protocol.Dns) {
	fmt.Println("writeNginxConf")

	filePath := constant.WorkDir + "/assets/nginxconf/" + dns.Spec.Host + ".conf"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("open file failed")
		fmt.Println(err.Error())
		return
	}
	defer file.Close()
	file.WriteString("server {\n")
	file.WriteString("    listen 80;\n")
	file.WriteString("    server_name " + dns.Spec.Host + ";\n")
	for _, p := range dns.Spec.Paths {
		file.WriteString("    location " + p.SubPath + " {\n")
		file.WriteString("        proxy_pass http://" + p.ServiceIp + ":" + strconv.Itoa(p.Port) + "/;\n")
		file.WriteString("    }\n")
	}
	file.WriteString("}\n")

}

func DeleteNginxConf(dns protocol.Dns) {
	fmt.Println("deleteNginxConf")
	filePath := constant.WorkDir + "/assets/nginxconf/" + dns.Spec.Host + ".conf"
	err := os.Remove(filePath)
	if err != nil {
		fmt.Println("delete file failed")
		fmt.Println(err.Error())
		return
	}
}
