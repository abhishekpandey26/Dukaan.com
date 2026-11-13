// Package utils has small helpers for reading and writing JSON over HTTP.
package utils

import (
	"encoding/json"
	"net/http"
)

// WriteJSON sends any Go value as a JSON response.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // headers must be set BEFORE calling WriteHeader
	json.NewEncoder(w).Encode(data)
}

// WriteError sends {"error": "message"}.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// ReadJSON decodes the request body into dst (dst must be a pointer).
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // limit body to 1 MB
	return json.NewDecoder(r.Body).Decode(dst)
}
