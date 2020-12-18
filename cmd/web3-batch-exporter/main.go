package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"web3-batch-exporter/internal/helper"
	"web3-batch-exporter/internal/worker/historical"
	"web3-batch-exporter/internal/worker/live"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var cancelChan = make(chan struct{})
var registry *prometheus.Registry
var router *mux.Router

func main() {
	router = mux.NewRouter()
	router.HandleFunc("/historical", HistoricalHandler).Methods("POST")
	router.HandleFunc("/live", LiveHandler).Methods("POST")
	serverPort := helper.GetEnv("SERVER_PORT")
	log.Fatal(http.ListenAndServe(":"+serverPort, router))
}

func HistoricalHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	urlParams := r.URL.Query()
	startBlockParam := urlParams.Get("startBlock") // e.g. 11460000
	endBlockParam := urlParams.Get("endBlock")     // e.g. 11460686

	if startBlockParam == "" {
		log.Fatal("Please provide a 'startBlock' via query param!")
	}
	if endBlockParam == "" {
		log.Fatal("Please provide an 'endBlock' via query param!")
	}

	startBlock, _ := strconv.Atoi(startBlockParam)
	endBlock, _ := strconv.Atoi(endBlockParam)

	go func() {
		historical.StreamDataToDB(body, startBlock, endBlock)
	}()
}

func LiveHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	registry = prometheus.NewPedanticRegistry()
	live.StartPolling(body, getCancelChan(), registry)
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}

func getCancelChan() chan struct{} {
	close(cancelChan)
	cancelChan = make(chan struct{})
	return cancelChan
}
