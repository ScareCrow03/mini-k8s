package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/message"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"net/http"
	"strconv"
	"time"

	"math/rand"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func GetClusterIP() string {
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	var cip string
	for {
		cip = "222.111." + strconv.Itoa(rand.Int()%256) + "." + strconv.Itoa(rand.Int()%256)
		reply, _ := st.Get(constant.EtcdServiceClusterIPPrefix + cip)
		if len(reply.Value) == 0 {
			break
		}
	}

	jsonstr, err := json.Marshal(cip)
	if err != nil {
		panic(err)
	}

	st.Put(constant.EtcdServiceClusterIPPrefix+cip, jsonstr)

	return cip
}

func DelClusterIP(cip string) string {
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	err = st.Del(constant.EtcdServiceClusterIPPrefix + cip)
	if err != nil {
		panic(err)
	}

	return cip
}

func CreateService(c *gin.Context) {
	var svc protocol.ServiceType
	c.BindJSON(&svc)
	if svc.Config.Metadata.Namespace == "" {
		svc.Config.Metadata.Namespace = "default"
	}
	data, _ := yaml.Marshal(svc)
	fmt.Printf("CreateService: %s\n", string(data))
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()

	svc.Config.Metadata.UID = "mini-k8s-service-" + uid.NewUid()
	svc.Config.Spec.ClusterIP = GetClusterIP()
	fmt.Println("CreateService cluster IP: ", svc.Config.Spec.ClusterIP)

	jsonstr, err := json.Marshal(svc)
	if err != nil {
		panic(err)
	}
	err = st.Put(constant.EtcdServicePrefix+svc.Config.Metadata.Namespace+"/"+svc.Config.Metadata.Name, jsonstr)
	if err != nil {
		panic(err)
	}

	message.Publish(message.CreateServiceQueueName, jsonstr)

	c.JSON(http.StatusOK, svc)
}

func DeleteService(c *gin.Context) {
	var svc protocol.ServiceType
	c.BindJSON(&svc)
	if svc.Config.Metadata.Namespace == "" {
		svc.Config.Metadata.Namespace = "default"
	}
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	// 从etcd里拿到这个service的clusterIP信息等，必须具备完全的信息，包括UID等！
	stored_svc_json, _ := st.Get(constant.EtcdServicePrefix + svc.Config.Metadata.Namespace + "/" + svc.Config.Metadata.Name)
	if len(stored_svc_json.Value) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"error": "service not found",
		})
		return
	}
	stored_svc := protocol.ServiceType{}
	err = json.Unmarshal(stored_svc_json.Value, &stored_svc)

	if err != nil {
		panic(err)
	}

	DelClusterIP(stored_svc.Config.Spec.ClusterIP)
	err = st.Del(constant.EtcdServicePrefix + svc.Config.Metadata.Namespace + "/" + svc.Config.Metadata.Name)
	if err != nil {
		panic(err)
	}

	message.Publish(message.DeleteServiceQueueName, stored_svc_json.Value)

	c.JSON(http.StatusOK, nil)
}

func GetAllServices() []protocol.ServiceType {
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		panic(err)
	}
	defer st.Close()
	reply, err := st.GetWithPrefix(constant.EtcdServicePrefix)
	if err != nil {
		panic(err)
	}
	var services []protocol.ServiceType
	for _, r := range reply {
		var s protocol.ServiceType
		err = json.Unmarshal(r.Value, &s)
		if err != nil {
			panic(err)
		}
		services = append(services, s)
	}
	return services
}

func ServiceCheckNow(c *gin.Context) {
	s := make(map[int]int)
	s[0] = 0
	jsonstr, _ := json.Marshal(s)
	message.Publish(message.ServiceCheckNowQueueName, jsonstr)
}

func CalculateServiceAndEps() {

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	st, _ := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	defer st.Close()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("CalculateServiceAndEps\n")
			st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
			if err != nil {
				panic(err)
			}
			defer st.Close()
			var pods []protocol.Pod
			reply, err := st.GetWithPrefix(constant.EtcdPodPrefix)
			if err != nil {
				panic(err)
			}
			for _, r := range reply {
				var p protocol.Pod
				err = json.Unmarshal(r.Value, &p)
				if err != nil {
					panic(err)
				}
				pods = append(pods, p)
			}

			reply2, err := st.GetWithPrefix(constant.EtcdServicePrefix)
			if err != nil {
				panic(err)
			}

			// 逐个计算eps
			for _, r := range reply2 {
				var s protocol.ServiceType
				err = json.Unmarshal(r.Value, &s)
				if err != nil {
					panic(err)
				}
				// 先清空旧的，重新计算一遍即可！
				s.Status.Endpoints = []protocol.Endpoint{}
				for _, p := range pods {
					if p.Status.IP == "" {
						continue
					}

					if protocol.IsSelectorMatchOnePodNoPointer(s.Config.Spec.Selector, p) {
						new_eps := protocol.GetEndpointsFromPod(&p)
						s.Status.Endpoints = append(s.Status.Endpoints, new_eps...)
					}
				}
				// 写回
				jsonstr, err := json.Marshal(s)
				if err != nil {
					panic(err)
				}
				st.Put(constant.EtcdServicePrefix+s.Config.Metadata.Namespace+"/"+s.Config.Metadata.Name, jsonstr)
			}
		}
	}

}
