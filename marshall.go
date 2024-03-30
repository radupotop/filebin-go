package main

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
	Context string `json:"context"`
	Status  int    `json:"status"`
}

// Return JSON response
func (r Response) returnJson(w http.ResponseWriter) {
	// Set the Content-Type header and the HTTP status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status)
	// Marshall struct to JSON
	json.NewEncoder(w).Encode(r)
}

// // MarshalJSON implements the json.Marshaler interface for Response
// func (r Response) MarshalJSON() ([]byte, error) {
// 	// Define a map to hold the message
// 	data := map[string]string{
// 		"message": r.Message,
// 		"status":  r.Status,
// 	}
// 	// Marshal the map into JSON
// 	return json.Marshal(data)
// }

// func returnJson(w http.ResponseWriter, resp Response) {
// 	json.NewEncoder(w).Encode(resp)
// }
