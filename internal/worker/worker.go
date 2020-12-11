package worker

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
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

func callWeb3BatchService(payload []byte) []byte {
	url := helper.GetEnv("WEB3_BATCH_SERVICE_URL")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func StartPolling(payload []byte, cancelChan chan struct{}, registry prometheus.Registerer) {
	dataMap := metric.GetDataMap()

	hasRegistered := false
	startTicker(cancelChan, func() {
		log.Println("calling web3-batch-service...")
		responseBytes := callWeb3BatchService(payload)
		log.Println("parsing result from web3-batch-service...")
		response := helper.ParseJSONResponse(responseBytes)
		log.Println("extracting data into dataMap...")
		metric.ExtractData(response, dataMap)
		if hasRegistered != true {
			log.Println("registering prom collectors...")
			prom.RegisterCollectors(dataMap, registry)
			hasRegistered = true
		}
		log.Println("ticker started.")
	})
}
