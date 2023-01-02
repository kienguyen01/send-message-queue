package main

import (
	"encoding/json"
	"log"
	"net/http"
	"errors"

	sender "github.com/kienguyen01/send-message-queue/sender"
	//receiver "github.com/kienguyen01/send-message-queue/receiver"
)

func main() {
	// Define a function that will be triggered by a POST request
	// This function just prints a message to the console
	triggeredFunction := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
	}

	// Create a handler for the "/trigger" endpoint
	http.HandleFunc("/send", triggeredFunction)

	// Start the server on port 8080
	http.ListenAndServe(":8080", nil)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func response(w http.ResponseWriter, message string, httpStatusCode int){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

func sendMessage(w http.ResponseWriter, r *http.Request){
	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		response(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	var m sender.Message
	var unmarshalErr *json.UnmarshalTypeError

	body := json.NewDecoder(r.Body)
	body.DisallowUnknownFields()

	err := body.Decode(&m)

	if err != nil {
		if errors.As(err, &unmarshalErr) {
			response(w, "Bad Request. Wrong Type provided for field "+unmarshalErr.Field, http.StatusBadRequest)
		} else {
			response(w, "Bad Request "+err.Error(), http.StatusBadRequest)
		}
	}
	sender.SendMessage(m)
}
