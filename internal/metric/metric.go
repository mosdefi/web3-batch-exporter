package metric

import (
	"math"
	"reflect"
	"strconv"

	"web3-batch-exporter/internal/helper"
)

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

func ExtractData(response map[string]interface{}) map[string]*[]MetricMap {
	data := make(map[string]*[]MetricMap)

	namespaces := extractNamespaces(response)
	for _, ns := range namespaces {
		nsMaps := []MetricMap{}
		data[ns] = &nsMaps

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
				//log.Println(k, v)
				metric := extractMetric(v)
				metric.Name = k
				metrics[k] = metric
			}
			scale(ns, metrics)
		}
	}
	return data
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
		//log.Println(v)
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
	//log.Println("Trying to extract value and type for", value)
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

func extractNamespaces(response map[string]interface{}) []string {
	var namespaces = []string{}
	for key := range response {
		namespaces = append(namespaces, key)
	}
	return namespaces
}
