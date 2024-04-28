package handler

import "mini-k8s/pkg/protocol"

func GetPodNode(podConfig protocol.PodConfig) string {
	return "node1"
}
