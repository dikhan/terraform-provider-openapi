package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pborman/uuid"
)

type ContentDeliveryNetwork struct {
	Slug      string   `json:"slug"`
	Label     string   `json:"label"`
	Ips       []string `json:"ips"`
	Hostnames []string `json:"hostnames"`
}

var db = map[string]*ContentDeliveryNetwork{}

func ContentDeliveryNetworkCreateV1(w http.ResponseWriter, r *http.Request) {
	cdn, err := readRequest(r)
	if err != nil {
		sendErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}
	cdn.Slug = uuid.New()
	db[cdn.Slug] = cdn
	sendResponse(http.StatusCreated, w, cdn)
}

func ContentDeliveryNetworkGetV1(w http.ResponseWriter, r *http.Request) {
	a, err := retrieveCdn(r)
	if err != nil {
		sendErrorResponse(http.StatusNotFound, err.Error(), w)
		return
	}
	sendResponse(http.StatusOK, w, a)
}

func ContentDeliveryNetworkUpdateV1(w http.ResponseWriter, r *http.Request) {
	a, err := retrieveCdn(r)
	if err != nil {
		sendErrorResponse(http.StatusNotFound, err.Error(), w)
		return
	}
	cdn, err := readRequest(r)
	if err != nil {
		sendErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}
	cdn.Slug = a.Slug
	db[cdn.Slug] = cdn
	sendResponse(http.StatusOK, w, cdn)
}

func ContentDeliveryNetworkDeleteV1(w http.ResponseWriter, r *http.Request) {
	a, err := retrieveCdn(r)
	if err != nil {
		sendErrorResponse(http.StatusNotFound, err.Error(), w)
		return
	}
	delete(db, a.Slug)
	updateResponseHeaders(http.StatusOK, w)
}

func retrieveCdn(r *http.Request) (*ContentDeliveryNetwork, error) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/cdns/")
	if id == "" {
		return nil, fmt.Errorf("cdn id path param not provided")
	}
	cdn := db[id]
	if cdn == nil {
		return nil, fmt.Errorf("cdn id '%s' not found", id)
	}
	return cdn, nil
}

func readRequest(r *http.Request) (*ContentDeliveryNetwork, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body - %s", err)
	}
	cdn := &ContentDeliveryNetwork{}
	if err := json.Unmarshal(body, cdn); err != nil {
		return nil, fmt.Errorf("payload does not match cdn spec - %s", err)
	}
	return cdn, nil
}

func sendResponse(httpResponseStatusCode int, w http.ResponseWriter, cdn *ContentDeliveryNetwork) {
	var resBody []byte
	var err error
	if resBody, err = json.Marshal(cdn); err != nil {
		msg := fmt.Sprintf("internal server error - %s", err)
		sendErrorResponse(http.StatusInternalServerError, msg, w)
	}
	w.WriteHeader(httpResponseStatusCode)
	w.Write(resBody)
}

func sendErrorResponse(httpStatusCode int, message string, w http.ResponseWriter) {
	updateResponseHeaders(httpStatusCode, w)
	w.Write([]byte(fmt.Sprintf(`{"code":"%d", "message": "%s"}`, httpStatusCode, message)))
}

func updateResponseHeaders(httpStatusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatusCode)
}
