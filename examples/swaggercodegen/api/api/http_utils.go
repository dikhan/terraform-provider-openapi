package api

import (
	"net/http"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
}

func sendErrorResponse(httpStatusCode int, message string, w http.ResponseWriter) {
	updateResponseHeaders(httpStatusCode, w)
	w.Write([]byte(fmt.Sprintf(`{"code":"%d", "message": "%s"}`, httpStatusCode, message)))
}

func updateResponseHeaders(httpStatusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatusCode)
}

