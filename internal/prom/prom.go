package prom

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"web3-batch-exporter/internal/metric"
)

type Web3SliceCollector struct {
	Web3Slice *Web3Slice
}

type Web3Slice struct {
	Namespace string
	Address   string
}

func NewWeb3Slice(namespace string, address string, registry prometheus.Registerer) *Web3Slice {
	w3slice := &Web3Slice{Namespace: namespace, Address: address}
	collector := Web3SliceCollector{Web3Slice: w3slice}
	dataMap := metric.GetDataMap()
	dataMap.AddCollector(collector)
	prometheus.WrapRegistererWith(nil, registry).MustRegister(collector)

	return w3slice
}

func (collector Web3SliceCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(collector, ch)
}

func (collector Web3SliceCollector) Collect(ch chan<- prometheus.Metric) {
	namespace := collector.Web3Slice.Namespace
	address := collector.Web3Slice.Address
	dataMap := metric.GetDataMap()
	metricMaps := dataMap.GetData(namespace)

	for _, metricMap := range *metricMaps {
		if metricMap.Address != address {
			continue
		}
		toExport := metric.FindExportableMetrics(metricMap.Metrics)
		for _, metric := range toExport {
			descId := namespace + "_" + address + "_" + metric.Name
			desc := prometheus.NewDesc(
				descId,
				"Auto-generated gauge by web3-batch-exporter.",
				[]string{"namespace", "address"},
				nil,
			)
			promMetric := prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				metric.Value.(float64),
				namespace,
				address,
			)

			seconds := int64(dataMap.BlockInfo.Timestamp) * 1000
			timestamp := time.Unix(0, seconds*int64(time.Second))
			ch <- prometheus.NewMetricWithTimestamp(
				timestamp,
				promMetric,
			)
		}
	}
}

func RegisterCollectors(dataMap *metric.DataMap, registry prometheus.Registerer) {
	unregisterCollectors(dataMap, registry)
	for _, namespace := range dataMap.GetNamespaces() {
		metricMaps := dataMap.GetData(namespace)
		for _, metricMap := range *metricMaps {
			NewWeb3Slice(namespace, metricMap.Address, registry)
		}
	}
}

func unregisterCollectors(dataMap *metric.DataMap, registry prometheus.Registerer) {
	for _, collector := range dataMap.GetCollectors() {
		registry.Unregister(collector)
	}
	dataMap.ResetCollectors()
}
