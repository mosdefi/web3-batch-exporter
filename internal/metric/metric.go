package metric

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"sync"

	"web3-batch-exporter/internal/helper"
	"web3-batch-exporter/internal/slice"

	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
)

var once sync.Once

var instance *DataMap

func GetDataMap() *DataMap {
	once.Do(func() {
		if instance == nil {
			instance = &DataMap{
				Data:       make(map[string]*[]MetricMap),
				Slices:     make(map[string]*slice.CSVSlice),
				Collectors: []prometheus.Collector{},
			}
		}
	})
	return instance
}

type DataMap struct {
	mux        sync.Mutex
	BlockInfo  *BlockInfo
	Data       map[string]*[]MetricMap
	Slices     map[string]*slice.CSVSlice
	Collectors []prometheus.Collector
}

type Metric struct {
	Name      string
	Value     interface{}
	ValueMap  map[string]interface{}
	ValueType string
}

type MetricMap struct {
	Address string
	Alias   string
	Metrics map[string]*Metric
}

type BlockInfo struct {
	Number    int64
	Timestamp float64
	Hash      string
	GasLimit  int64
	GasUsed   int64
}

func (dataMap *DataMap) AddToSlice(key string, row string, format string) *slice.CSVSlice {
	dataMap.mux.Lock()
	defer dataMap.mux.Unlock()
	s, ok := dataMap.Slices[key]
	if !ok {
		s = &slice.CSVSlice{
			Format: format,
			Rows:   []string{},
		}
		dataMap.Slices[key] = s
	}
	s.Rows = append(s.Rows, row)
	return s
}

func (dataMap *DataMap) GetData(key string) *[]MetricMap {
	dataMap.mux.Lock()
	defer dataMap.mux.Unlock()
	return dataMap.Data[key]
}

func (dataMap *DataMap) SetData(key string, data *[]MetricMap, blockInfo *BlockInfo) {
	dataMap.mux.Lock()
	dataMap.Data[key] = data
	dataMap.BlockInfo = blockInfo
	dataMap.mux.Unlock()
}

func (dataMap *DataMap) GetNamespaces() []string {
	dataMap.mux.Lock()
	defer dataMap.mux.Unlock()

	var namespaces = []string{}
	for key, _ := range dataMap.Data {
		namespaces = append(namespaces, key)
	}
	return namespaces
}

func (dataMap *DataMap) GetCollectors() []prometheus.Collector {
	dataMap.mux.Lock()
	defer dataMap.mux.Unlock()
	return dataMap.Collectors
}

func (dataMap *DataMap) AddCollector(collector prometheus.Collector) {
	dataMap.mux.Lock()
	dataMap.Collectors = append(dataMap.Collectors, collector)
	dataMap.mux.Unlock()
}

func (dataMap *DataMap) ResetCollectors() {
	dataMap.mux.Lock()
	dataMap.Collectors = []prometheus.Collector{}
	dataMap.mux.Unlock()
}

func (dataMap *DataMap) ToCSVSlices() {
	namespaces := dataMap.GetNamespaces()
	for _, namespace := range namespaces {
		metricMaps := dataMap.GetData(namespace)

		for _, metricMap := range *metricMaps {
			address := metricMap.Address
			key := namespace + "#" + address

			toExport := FindExportableMetrics(metricMap.Metrics)
			row := fmt.Sprintf("%s,%s,%s", namespace, address, strconv.FormatFloat(dataMap.BlockInfo.Timestamp, 'f', -1, 64))

			// https://github.com/VictoriaMetrics/VictoriaMetrics#how-to-import-csv-data
			format := "1:label:namespace,2:label:address,3:time:unix_s"
			formatIdx := 4
			for _, metric := range toExport {
				format += fmt.Sprintf(",%d:metric:%s", formatIdx, metric.Name)
				row += fmt.Sprintf(",%f", metric.Value.(float64))
				formatIdx += 1
			}

			slice := dataMap.AddToSlice(key, row, format)
			slice.SendInBatch()
		}
	}
}

func ExtractData(responseBytes []byte, dataMap *DataMap) {
	//log.Println("parsing result from web3-batch-service...")
	response := helper.ParseJSONResponse(responseBytes)
	namespaces, blockInfo := extractNamespacesAndBlockInfo(response)
	for _, ns := range namespaces {
		nsMaps := []MetricMap{}
		dataMap.SetData(ns, &nsMaps, &blockInfo)

		values := response[ns]
		for _, entry := range values.([]interface{}) {
			address, ok := entry.(map[string]interface{})["address"].(string)
			if !ok {
				panic("No address found in result.")
			}

			// metrics for a single address
			metrics := make(map[string]*Metric)
			// a map holding metrics for all addresses
			metricMap := MetricMap{Address: address, Metrics: metrics}
			// a map holding metricMaps for all namespaces
			nsMaps = append(nsMaps, metricMap)

			for k, v := range entry.(map[string]interface{}) {
				metric := extractMetric(v)
				metric.Name = k
				metrics[k] = metric
			}
			scale(ns, metrics)
		}
	}
}

func FindExportableMetrics(metricMap map[string]*Metric) []*Metric {
	toExport := []*Metric{}
	for _, metric := range metricMap {
		if metric.ValueType == "float64" {
			toExport = append(toExport, metric)
		}
	}
	return toExport
}

func scale(namespace string, metrics map[string]*Metric) {
	decimals, ok := metrics["decimals"]
	if ok {
		scaleMetrics := findScaleMetrics(namespace, metrics)
		for _, m := range scaleMetrics {
			fv := m.Value.(float64)
			dv := int(decimals.Value.(float64))
			metric := metrics[m.Name]
			metric.Value = fv / math.Pow10(dv)
		}
	}
}

func findScaleMetrics(namespace string, metricMap map[string]*Metric) []*Metric {
	scaleMetrics := []*Metric{}
	toExport := FindExportableMetrics(metricMap)
	for _, metric := range toExport {
		if helper.InList(metric.Name, getScaleList(namespace)) {
			scaleMetrics = append(scaleMetrics, metric)
		}
	}
	return scaleMetrics
}

func getScaleList(namespace string) []string {
	scaleList := []string{"getPricePerFullShare", "balanceOf", "available", "balance"}
	// TODO get this from some kind of map or so per namespace
	return scaleList
}

func extractMetric(value interface{}) *Metric {
	var metric = Metric{}
	switch v := value.(type) {
	case []interface{}:
		first := v[0] // TODO implement all results not only the first one
		inner := first.(map[string]interface{})["value"]

		switch innerTyped := inner.(type) {
		case map[string]interface{}:
			metric.ValueMap = innerTyped //TODO implement nested results
		default:
			metric.Value, metric.ValueType = extractValueAndType(inner)
		}
	default:
		metric.Value, metric.ValueType = extractValueAndType(v)
	}

	return &metric
}

func extractValueAndType(value interface{}) (interface{}, string) {
	var floatType = reflect.TypeOf(float64(0))
	var intType = reflect.TypeOf(int(0))
	var stringType = reflect.TypeOf("")
	var boolType = reflect.TypeOf(false)
	v := reflect.ValueOf(value)
	v = reflect.Indirect(v)

	if v.Type().ConvertibleTo(floatType) {
		fv := v.Convert(floatType)
		return fv.Float(), "float64"

	} else if v.Type() == intType {
		return float64(v.Int()), "float64"

	} else if v.Type() == stringType {
		fv, err := strconv.ParseFloat(v.String(), 64)
		if err == nil {
			return fv, "float64"
		}

	} else if v.Type() == boolType {
		if (v.Bool()) == true {
			return 1.0, "float64"
		}
		return 0.0, "float64"
	}

	return v.String(), "string"
}

func extractNamespacesAndBlockInfo(response map[string]interface{}) ([]string, BlockInfo) {
	var namespaces = []string{}
	var blockInfo BlockInfo
	for key := range response {
		if key == "blockInfo" {
			err := mapstructure.Decode(response[key], &blockInfo)
			if err != nil {
				log.Println(err)
			}
		} else {
			namespaces = append(namespaces, key)
		}
	}
	return namespaces, blockInfo
}
