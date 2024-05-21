package main

import "mini-k8s/pkg/controller"

func main() {
	// var replicasetController controller.ReplicasetController
	// replicasetController.Run()
	// controller.Init()
	var replicasetController controller.ReplicasetController
	go replicasetController.Start()

	var hpaController controller.HPAController
	go hpaController.Start()

	var PingSourceController controller.PingSourceController
	go PingSourceController.Start()
	select {}
}
