package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tsawler/toolbox"
)

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := toolbox.JSONResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	var tools toolbox.Tools

	_ = tools.WriteJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	var tools toolbox.Tools

	err := tools.ReadJSON(w, r, &requestPayload)
	if err != nil {
		tools.ErrorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	default:
		tools.ErrorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")
	var tools toolbox.Tools

	// call the service
	request, err := http.NewRequest(
		"POST",
		"http://authentication-service/authenticate",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		tools.ErrorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		tools.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// create a variable we'll read response.Body into
	var jsonFromService toolbox.JSONResponse

	// decode json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		tools.ErrorJSON(w, err)
		return
	}

	// guard for authentication issues
	if response.StatusCode != http.StatusAccepted || jsonFromService.Error {
		tools.ErrorJSON(w, errors.New(jsonFromService.Message), http.StatusUnauthorized)
		return
	}

	payload := toolbox.JSONResponse{
		Error:   false,
		Message: "Authenticated!",
		Data:    jsonFromService.Data,
	}

	tools.WriteJSON(w, http.StatusAccepted, payload)
}
