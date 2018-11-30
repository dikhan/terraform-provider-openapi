package openapi

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func prettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	log.Printf(string(b))
	log.Println()
}

func sPrettyPrint(v interface{}) string {
	if v == nil {
		return "nil"
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func getFileContent(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
