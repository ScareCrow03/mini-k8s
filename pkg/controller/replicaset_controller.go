package controller

import "time"

type ReplicasetController struct {
}

func (rsc *ReplicasetController) CheckReplicaset() {

}

func (rsc *ReplicasetController) Run() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				rsc.CheckReplicaset()
			}
		}
	}()
}
