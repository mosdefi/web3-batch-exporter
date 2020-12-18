package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func InList(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GetEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("Error ENV variable %s is not set.\n", key)
	}
	return val
}

func CallWeb3BatchService(payload []byte, blockNum int) []byte {
	url := GetEnv("WEB3_BATCH_SERVICE_URL")

	if blockNum > 0 {
		url += fmt.Sprintf("?blockNum=%d", blockNum)
	}
	log.Println(url)

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	return doPost(payload, url, headers)
}

func PushToTimeseriesDB(format string, rows []string) {
	url := GetEnv("VICTORIA_METRICS_URL")
	url += "?format=" + format
	payload := []byte(strings.Join(rows, "\n"))
	doPost(payload, url, nil)
	log.Printf("Successfully pushed %d rows to VM", len(rows))
}

func doPost(payload []byte, url string, headers map[string]string) []byte {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if headers != nil {
		for header, val := range headers {
			req.Header.Set(header, val)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func ParseJSONResponse(bytes []byte) map[string]interface{} {
	var response map[string]interface{}
	if err := json.Unmarshal(bytes, &response); err != nil {
		log.Fatal(err)
	}

	return response
}
