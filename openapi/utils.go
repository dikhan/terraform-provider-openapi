package openapi

import (
	"encoding/json"
	"log"
)

func prettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	log.Printf(string(b))
	log.Println()
}

func responseContainsExpectedStatus(expectedStatusCodes []int, responseStatusCode int) bool {
	for _, expectedStatusCode := range expectedStatusCodes {
		if expectedStatusCode == responseStatusCode {
			return true
		}
	}
	return false
}
