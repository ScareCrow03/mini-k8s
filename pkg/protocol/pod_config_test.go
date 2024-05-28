package protocol

import (
	"fmt"
	yamlParse "mini-k8s/pkg/utils/yaml"
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestYAMLToPodConfig(t *testing.T) {
	var podConfig PodConfig
	//podConfig.YAMLToPodConfig("../../assets/pod_config_test1.yaml")
	yamlParse.YAMLParse(&podConfig, "../../assets/pod_config_test1.yaml")
	fmt.Println("ApiVersion: ", podConfig.ApiVersion)
	fmt.Println("Kind: ", podConfig.Kind)
	fmt.Println("Metadata: ")
	fmt.Println("    Name: ", podConfig.Metadata.Name)
	fmt.Println("    Namespace: ", podConfig.Metadata.Namespace)
	fmt.Println("    Labels: ")
	for k, v := range podConfig.Metadata.Labels {
		fmt.Println("        ", k, ": ", v)
	}
	fmt.Println("    Annotations: ")
	for k, v := range podConfig.Metadata.Annotations {
		fmt.Println("        ", k, ": ", v)
	}
	fmt.Println("Spec: ")
	fmt.Println("    RestartPolicy: ", podConfig.Spec.RestartPolicy)
	fmt.Println("    Containers: ")
	for i, v := range podConfig.Spec.Containers {
		fmt.Println("        Index: ", i)
		fmt.Println("        Name: ", v.Name)
		fmt.Println("        Image: ", v.Image)
		fmt.Println("        VolumeMounts: ")
		for i1, v1 := range v.VolumeMounts {
			fmt.Println("            Index: ", i1)
			fmt.Println("            Name: ", v1.Name)
			fmt.Println("            MountPath: ", v1.MountPath)
		}
	}
	fmt.Println("    NodeSelector: ")
	for k, v := range podConfig.Spec.NodeSelector {
		fmt.Println("        ", k, ": ", v)
	}
	fmt.Println("    Volumes: ")
	for i, v := range podConfig.Spec.Volumes {
		fmt.Println("        index: ", i)
		fmt.Println("        Name: ", v.Name)
		fmt.Println("        HostPath.Path: ", v.HostPath.Path)
		fmt.Println("        HostPath.Type: ", v.HostPath.Type)
	}
}
