package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pborman/uuid"
	"log"
)

var db = map[string]*ContentDeliveryNetwork{}

func ContentDeliveryNetworkCreateV1(w http.ResponseWriter, r *http.Request) {
	if AuthenticateRequest(r, w) != nil {
		return
	}
	xRequestID := r.Header.Get("X-Request-ID")
	log.Printf("Header [X-Request-ID]: %s", xRequestID)
	cdn := &ContentDeliveryNetwork{}
	err := readRequest(r, cdn)
	if err != nil {
		sendErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}
	cdn.Id = uuid.New()
	db[cdn.Id] = cdn
	sendResponse(http.StatusCreated, w, cdn)
}

func ContentDeliveryNetworkGetV1(w http.ResponseWriter, r *http.Request) {
	if AuthenticateRequest(r, w) != nil {
		return
	}
	cdn, err := retrieveCdn(r)
	if err != nil {
		sendErrorResponse(http.StatusNotFound, err.Error(), w)
		return
	}
	sendResponse(http.StatusOK, w, cdn)
}

func ContentDeliveryNetworkUpdateV1(w http.ResponseWriter, r *http.Request) {
	if AuthenticateRequest(r, w) != nil {
		return
	}
	cdn, err := retrieveCdn(r)
	if err != nil {
		sendErrorResponse(http.StatusNotFound, err.Error(), w)
		return
	}
	newCDN := &ContentDeliveryNetwork{}
	err = readRequest(r, newCDN)
	if err != nil {
		sendErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}
	newCDN.Id = cdn.Id
	db[cdn.Id] = newCDN
	sendResponse(http.StatusOK, w, newCDN)
}

func ContentDeliveryNetworkDeleteV1(w http.ResponseWriter, r *http.Request) {
	if AuthenticateRequest(r, w) != nil {
		return
	}
	cdn, err := retrieveCdn(r)
	if err != nil {
		sendErrorResponse(http.StatusNotFound, err.Error(), w)
		return
	}
	delete(db, cdn.Id)
	sendResponse(http.StatusNoContent, w, nil)
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