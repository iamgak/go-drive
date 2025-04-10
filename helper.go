package main

import (
	"encoding/json"
	"net/http"
)

func (app *Application) ServerError(w http.ResponseWriter, err error) {
	app.Logger.Error("Internal Server Error: ", err)
	app.sendJSONResponse(w, http.StatusInternalServerError, "Internal Server Error")
}

func (app *Application) CustomError(w http.ResponseWriter, status int, msg string) {
	app.Logger.Error("Internal Server Error: ", msg)
	app.ErrorJSONResponse(w, status, msg)
}

func (app *Application) sendJSONResponse(w http.ResponseWriter, statusCode int, message any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := map[string]interface{}{
		"status":  true,
		"message": message,
	}

	json.NewEncoder(w).Encode(resp)
}

func (app *Application) ErrorJSONResponse(w http.ResponseWriter, statusCode int, message any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := map[string]interface{}{
		"status": false,
		"error":  message,
	}

	json.NewEncoder(w).Encode(resp)
}
