package worker

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"web3-batch-exporter/internal/helper"
	"web3-batch-exporter/internal/metric"
	"web3-batch-exporter/internal/prometheus"
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

func StartPolling(payload []byte, cancelChan chan struct{}) {
	startTicker(cancelChan, func() {
		fmt.Println("calling web3-batch-service...")
		responseBytes := callWeb3BatchService(payload)
		response := helper.ParseJSONResponse(responseBytes)
		data := metric.ExtractData(response)
		prometheus.PushToPrometheus(data)
	})
}
