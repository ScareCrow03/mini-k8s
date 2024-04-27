package protocol

import (
	"fmt"
	"mini-k8s/pkg/utils/yaml"
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestYAMLToPodConfig(t *testing.T) {
	var podConfig PodConfig
	//podConfig.YAMLToPodConfig("../../assets/pod_config_test1.yaml")
	yamlParse.YAMLParse(&podConfig, "../../assets/pod_config_test1.yaml")
	fmt.Print("ApiVersion: ", podConfig.ApiVersion, "\n")
	fmt.Print("Kind: ", podConfig.Kind, "\n")
	fmt.Print("Metadata: ", "\n")
	fmt.Print("    Name: ", podConfig.Metadata.Name, "\n")
	fmt.Print("    Namespace: ", podConfig.Metadata.Namespace, "\n")
	fmt.Print("    Labels: ", "\n")
	for k, v := range podConfig.Metadata.Labels {
		fmt.Print("        ", k, ": ", v, "\n")
	}
	fmt.Print("    Annotations: ", "\n")
	for k, v := range podConfig.Metadata.Annotations {
		fmt.Print("        ", k, ": ", v, "\n")
	}
	fmt.Print("Spec: ", "\n")
	fmt.Print("    RestartPolicy: ", podConfig.Spec.RestartPolicy, "\n")
	fmt.Print("    Containers: ", "\n")
	for i, v := range podConfig.Spec.Containers {
		fmt.Print("        Index: ", i, "\n")
		fmt.Print("        Name: ", v.Name, "\n")
		fmt.Print("        Image: ", v.Image, "\n")
		fmt.Print("        VolumeMounts: ", "\n")
		for i1, v1 := range v.VolumeMounts {
			fmt.Print("            Index: ", i1, "\n")
			fmt.Print("            Name: ", v1.Name, "\n")
			fmt.Print("            MountPath: ", v1.MountPath, "\n")
		}
	}
	fmt.Print("    NodeSelector: ", "\n")
	for k, v := range podConfig.Spec.NodeSelector {
		fmt.Print("        ", k, ": ", v, "\n")
	}
	fmt.Print("    Volumes: ", "\n")
	for i, v := range podConfig.Spec.Volumes {
		fmt.Print("        index: ", i, "\n")
		fmt.Print("        Name: ", v.Name, "\n")
		fmt.Print("        HostPath.Path: ", v.HostPath.Path, "\n")
		fmt.Print("        HostPath.Type: ", v.HostPath.Type, "\n")
	}
}
