package controller

import (
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/httputils"
	"mini-k8s/pkg/logger"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/type_cast"
	"time"

	"encoding/json"

	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

type PingSourceController struct {
	PingSourcesMap map[string]*protocol.CRType
	Crons          map[string]*cron.Cron
	QuitChs        map[string]chan struct{}
}

func (psc *PingSourceController) Start() {
	psc.PingSourcesMap = make(map[string]*protocol.CRType)
	psc.Crons = make(map[string]*cron.Cron)
	psc.QuitChs = make(map[string]chan struct{})
	ticker := time.NewTicker(15 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				psc.CheckAllPingSources()
			}
		}
	}()
}

func (psc *PingSourceController) CheckAllPingSources() {
	fmt.Printf("CheckAllPingSources\n")
	req, _ := json.Marshal("PingSource")
	resp := httputils.Post(constant.HttpPreffix+"/getObjectByType", req)
	var pingSources []protocol.CRType
	err := json.Unmarshal(resp, &pingSources)
	if err != nil {
		logger.KError("unmarshal pingSources error %s", err)
		return
	}

	updatedPingSources := make(map[string]*protocol.CRType)
	for _, ps := range pingSources {
		psKey := ps.Metadata.UID
		updatedPingSources[psKey] = &ps

		if _, ok := psc.PingSourcesMap[psKey]; !ok {
			data, _ := yaml.Marshal(ps)
			fmt.Printf("Detect new PingSource created, create cron for it: \n%s\n,", string(data))
			// 按cron表达式，建立定时任务，保存在这个定时器协程里
			var psSpec protocol.PingSourceSpec
			err = type_cast.GetObjectFromInterface(ps.Spec, &psSpec)
			if err != nil {
				logger.KError("Failed to get PingSourceSpec from PingSource: %s", err)
				continue
			}
			c := cron.New(cron.WithSeconds())
			// 配置可退出的定时器函数
			_, err := c.AddFunc(psSpec.Schedule, func() {
				select {
				case <-psc.QuitChs[psKey]:
					return
				default:
					psc.CheckOnePingSource(ps)
				}
			})
			if err != nil {
				logger.KError("Failed to create cron for PingSource: %s", err)
				continue
			}
			c.Start()
			psc.Crons[psKey] = c
			psc.QuitChs[psKey] = make(chan struct{})
			psc.PingSourcesMap[psKey] = &ps
		}
	}

	for psKey, c := range psc.Crons {
		if _, ok := updatedPingSources[psKey]; !ok {
			data, _ := yaml.Marshal(psc.PingSourcesMap[psKey])
			fmt.Printf("Detect current PingSource removed, stop cron for it: \n%s\n", string(data))
			c.Stop()
			close(psc.QuitChs[psKey])
			delete(psc.Crons, psKey)
			delete(psc.QuitChs, psKey)
			delete(psc.PingSourcesMap, psKey)
		}
	}
}

func (psc *PingSourceController) CheckOnePingSource(ps protocol.CRType) {
	fmt.Printf("CheckOnePingSource %s/%s\n", ps.Metadata.Namespace, ps.Metadata.Name)
	// 这里可以添加写入消息队列的逻辑，暂时不管它
}
