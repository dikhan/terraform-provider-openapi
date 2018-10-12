package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func readRequest(r *http.Request, in interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body - %s", err)
	}
	if err := json.Unmarshal(body, in); err != nil {
		return fmt.Errorf("payload does not match cdn spec - %s", err)
	}
	return nil
}

func sendResponse(httpResponseStatusCode int, w http.ResponseWriter, out interface{}) {
	var resBody []byte
	var err error
	if out != nil {
		if resBody, err = json.Marshal(out); err != nil {
			msg := fmt.Sprintf("internal server error - %s", err)
			sendErrorResponse(http.StatusInternalServerError, msg, w)
		}
	}
	updateResponseHeaders(httpResponseStatusCode, w)
	if len(resBody) > 0 {
		w.Write(resBody)
	}
	log.Printf("Response sent '%+v'", out)
}

func sendErrorResponse(httpStatusCode int, message string, w http.ResponseWriter) {
	updateResponseHeaders(httpStatusCode, w)
	err := fmt.Sprintf(`{"code":"%d", "message": "%s"}`, httpStatusCode, message)
	w.Write([]byte(err))
	log.Printf("Error Response sent '%s'", err)
}

func updateResponseHeaders(httpStatusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatusCode)
}
