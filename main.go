package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os",
	"net/http"

	sender "github.com/kienguyen01/send-message-queue/sender"
	// receiver "github.com/kienguyen01/send-message-queue/receiver"
)

func main() {
	fmt.Println("started listening to trigger")

	// Create a handler for the "/trigger" endpoint
	http.HandleFunc("/Send", func(w http.ResponseWriter, r *http.Request) {
		sendMessage(w, r)

		// Send a response
		fmt.Fprintf(w, "Function triggered successfully")
	})

	http.HandleFunc("/SendMultiple", func(w http.ResponseWriter, r *http.Request) {
		sendMultipleMessage(w, r)

		// Send a response
		fmt.Fprintf(w, "Function triggered successfully")
	})

	// Start the server on port 8080
	http.ListenAndServe(":8080", nil)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func response(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

func processMessage(w http.ResponseWriter, r *http.Request, decodeFunc func(*json.Decoder) error) {
	headerContentType := r.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		response(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	body := json.NewDecoder(r.Body)
	//body.DisallowUnknownFields()

	err := decodeFunc(body)
	if err != nil {
		var unmarshalErr *json.UnmarshalTypeError
		if errors.As(err, &unmarshalErr) {
			response(w, "Bad Request. Wrong Type provided for field "+unmarshalErr.Field, http.StatusBadRequest)
		} else {
			response(w, "Bad Request "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	response(w, "Sent messages", http.StatusOK)
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	processMessage(w, r, func(decoder *json.Decoder) error {
		var m sender.Message
		return decoder.Decode(&m)
	})
	sender.SendMessage(m)
}

func sendMultipleMessage(w http.ResponseWriter, r *http.Request) {
	processMessage(w, r, func(decoder *json.Decoder) error {
		var m sender.MultipleReceiverMessage
		return decoder.Decode(&m)
	})
	sender.SendMultipleMessages(m)
}