package live

import (
	"log"
	"time"
	"web3-batch-exporter/internal/helper"
	"web3-batch-exporter/internal/metric"
	"web3-batch-exporter/internal/prom"

	"github.com/prometheus/client_golang/prometheus"
)

func startTicker(cancelChan chan struct{}, f func()) {
	go func() {
		ticker := time.NewTicker(time.Second * 60)
		defer ticker.Stop()
	loop:
		for {
			f()
			select {
			case <-ticker.C:
				continue
			case <-cancelChan:
				break loop
			}
		}
	}()
}

func StartPolling(payload []byte, cancelChan chan struct{}, registry prometheus.Registerer) {
	dataMap := metric.GetDataMap()

	hasRegistered := false
	startTicker(cancelChan, func() {
		log.Println("calling web3-batch-service...")
		responseBytes := helper.CallWeb3BatchService(payload, 0)
		log.Println("extracting data into dataMap...")
		metric.ExtractData(responseBytes, dataMap)
		if hasRegistered != true {
			log.Println("registering prom collectors...")
			prom.RegisterCollectors(dataMap, registry)
			hasRegistered = true
		}
		log.Println("ticker started.")
	})
}
