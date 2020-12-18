package historical

import (
	"log"
	"runtime"
	"sync"
	"web3-batch-exporter/internal/helper"
	"web3-batch-exporter/internal/metric"
)

func StreamDataToDB(payload []byte, startBlock int, endBlock int) {
	if startBlock > endBlock {
		log.Fatal("startBlock must be before endBlock")
	}
	numWorkers := runtime.NumCPU()
	numBlocks := endBlock - startBlock + 1
	blockNumbers := make(chan int, numBlocks)
	dataMap := metric.GetDataMap()

	var wg sync.WaitGroup

	for id := 0; id < numWorkers; id++ {
		wg.Add(1)
		go worker(id, payload, blockNumbers, &wg)
	}

	for b := startBlock; b <= endBlock; b++ {
		blockNumbers <- b
	}
	close(blockNumbers)
	wg.Wait()

	for _, slice := range dataMap.Slices {
		slice.SendRemainder()
	}
}

func worker(workerID int, payload []byte, blockNumbers chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	dataMap := metric.GetDataMap()

	for blockNum := range blockNumbers {
		log.Println("workerID", workerID)
		log.Println("requesting web3 data for blockNum", blockNum)
		responseBytes := helper.CallWeb3BatchService(payload, blockNum)
		metric.ExtractData(responseBytes, dataMap)
		dataMap.ToCSVSlices()
	}
}
