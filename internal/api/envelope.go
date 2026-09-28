package api

import (
	"encoding/json"
	"net/http"
)

// Envelope wraps every successful response body in a uniform shape.
type Envelope struct {
	Data any `json:"data"`
}

// WriteData writes data to w, status code status, wrapped in an Envelope.
func WriteData(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(Envelope{Data: data})
}
