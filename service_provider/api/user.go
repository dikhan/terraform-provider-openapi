package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type User struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Phone     string `json:"phone"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	u, err := readRequest(w, r)
	if err != nil {
		return
	}
	sendResponse(w, u)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func GetUserByName(w http.ResponseWriter, r *http.Request) {
	u := &User{
		Username:  "dikhan",
		FirstName: "Daniel",
		LastName:  "Khan",
		Email:     "info@server.com",
		Phone:     "6049991234",
	}
	sendResponse(w, u)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/v2/users/")
	if username == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf(`{"msg":"username not found"}`)))
	}
	u, err := readRequest(w, r)
	if err != nil {
		return
	}
	u.Username = username
	sendResponse(w, u)
}

func readRequest(w http.ResponseWriter, r *http.Request) (*User, error) {
	body, err := ioutil.ReadAll(r.Body)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"msg":"failed to read request body", "error": "%s"}`, err)))
		return nil, err
	}
	u := &User{}
	if err := json.Unmarshal(body, u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"msg":"payload does not match user spec - req: %s", "error": "%s"}`, string(body), err)))
		return nil, err
	}
	return u, nil
}

func sendResponse(w http.ResponseWriter, u *User) {
	var resBody []byte
	var err error
	if resBody, err = json.Marshal(u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"msg":"internal server error","error": "%s"}`, err)))
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resBody)
}
