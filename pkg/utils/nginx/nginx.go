package nginx

import (
	"fmt"
	"mini-k8s/pkg/protocol"
	"os"
	"strconv"
)

func WriteNginxConf(dns protocol.Dns) {
	fmt.Println("writeNginxConf")

	filePath := "/home/lrh/Desktop/mini-k8s/assets/" + dns.Spec.Host + ".conf"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
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
		file.WriteString("        proxy_pass http://" + p.ServiceIp + ":" + strconv.Itoa(p.Port) + ";\n")
		file.WriteString("    }\n")
	}
	file.WriteString("}\n")

}
