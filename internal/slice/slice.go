package slice

import (
	"sync"
	"web3-batch-exporter/internal/helper"
)

type CSVSlice struct {
	mux    sync.Mutex
	Format string
	Rows   []string
}

func (slice *CSVSlice) SendInBatch() {
	slice.mux.Lock()
	maxSliceSize := 100
	if len(slice.Rows) == maxSliceSize {
		slice.send()
	}
	slice.mux.Unlock()
}

func (slice *CSVSlice) SendRemainder() {
	slice.mux.Lock()
	if len(slice.Rows) > 0 {
		slice.send()
	}
	slice.mux.Unlock()
}

func (slice *CSVSlice) send() {
	helper.PushToTimeseriesDB(slice.Format, slice.Rows)
	slice.Rows = []string{}
}
