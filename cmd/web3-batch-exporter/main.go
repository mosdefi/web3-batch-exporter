package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"web3-batch-exporter/internal/helper"
	"web3-batch-exporter/internal/worker"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var cancelChan = make(chan struct{})
var registry *prometheus.Registry
var router *mux.Router

func JSONHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	registry = prometheus.NewPedanticRegistry()
	worker.StartPolling(body, getCancelChan(), registry)
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}

func getCancelChan() chan struct{} {
	close(cancelChan)
	cancelChan = make(chan struct{})
	return cancelChan
}

func main() {
	router = mux.NewRouter()
	router.HandleFunc("/", JSONHandler).Methods("POST")
	serverPort := helper.GetEnv("SERVER_PORT")
	log.Fatal(http.ListenAndServe(":"+serverPort, router))
}
