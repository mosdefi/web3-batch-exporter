package helper

import (
	"encoding/json"
	"log"
	"os"
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

func ParseJSONResponse(bytes []byte) map[string]interface{} {
	/*jsonFile, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}*/

	var response map[string]interface{}
	if err := json.Unmarshal(bytes, &response); err != nil {
		log.Fatal(err)
	}

	return response
}
