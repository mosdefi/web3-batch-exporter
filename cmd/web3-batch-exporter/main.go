package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"web3-batch-exporter/internal/helper"
	"web3-batch-exporter/internal/worker"

	"github.com/gorilla/mux"
)

var cancelChan = make(chan struct{})

func JSONHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	worker.StartPolling(body, getCancelChan())
}

func getCancelChan() chan struct{} {
	close(cancelChan)
	cancelChan = make(chan struct{})
	return cancelChan
}

func main() {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", JSONHandler).Methods("POST")

	// Bind to a port and pass our router in
	serverPort := helper.GetEnv("SERVER_PORT")
	log.Fatal(http.ListenAndServe(":"+serverPort, r))
}
