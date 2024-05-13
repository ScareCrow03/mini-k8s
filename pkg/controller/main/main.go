package main

import "mini-k8s/pkg/controller"

func main() {
	// var replicasetController controller.ReplicasetController
	// replicasetController.Run()
	controller.Init()
	var replicasetController controller.ReplicasetController
	replicasetController.Start()
	select {}
}
