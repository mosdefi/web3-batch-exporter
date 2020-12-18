module web3-batch-exporter

go 1.15

replace web3-batch-exporter/internal/helper => ./internal/helper

replace web3-batch-exporter/internal/metric => ./internal/metric

replace web3-batch-exporter/internal/slice => ./internal/slice

replace web3-batch-exporter/internal/prom => ./internal/prom

replace web3-batch-exporter/internal/worker/live => ./internal/worker/live

replace web3-batch-exporter/internal/worker/historical => ./internal/worker/historical


require (
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/mitchellh/mapstructure v1.4.0
	github.com/prometheus/client_golang v1.8.0
)
