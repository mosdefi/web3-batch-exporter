package prometheus

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/push"

	"web3-batch-exporter/internal/helper"
	"web3-batch-exporter/internal/metric"
)

func PushToPrometheus(data map[string]*[]metric.MetricMap) {
	registry := prometheus.NewRegistry()
	factory := promauto.With(registry)
	pushGwURL := helper.GetEnv("PUSH_GW_URL")
	pusher := push.New(pushGwURL, "web3-batch-exporter").Gatherer(registry)

	var numExported int
	for ns, metricMaps := range data {
		for _, metricMap := range *metricMaps {
			toExport := metric.FindExportableMetrics(metricMap.Metrics)
			numExported += len(toExport)

			label := prometheus.Labels{"namespace": ns, "address": metricMap.Address}

			for _, metric := range toExport {
				//log.Println(ns, metric.Name, metric.Value)
				gauge := factory.NewGaugeVec(
					prometheus.GaugeOpts{
						Namespace: ns,
						Subsystem: metricMap.Address,
						Name:      metric.Name,
						Help:      "Auto-generated gauge by web3-batch-exporter.",
					},
					[]string{"namespace", "address"},
				)
				gauge.With(label).Set(metric.Value.(float64))
			}
		}
	}
	if err := pusher.Add(); err != nil {
		log.Fatal("Could not push to Pushgateway:", err)
	}

	log.Printf("Successfully exported %v metrics to prometheus push gateway.", numExported)
}
