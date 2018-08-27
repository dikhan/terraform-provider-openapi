package api

import (
	"net/http"
	"fmt"
	"errors"
)

func AuthenticateRequest(r *http.Request, w http.ResponseWriter) error {
	apiKey := r.Header.Get("Authorization")
	if apiKey != "apiKeyValue" {
		msg := fmt.Sprintf("unauthorized user")
		sendErrorResponse(http.StatusUnauthorized, msg, w)
		return errors.New(msg)
	}
	return nil
}