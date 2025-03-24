package utils

import (
	"encoding/json"
	"net/http"
)

// RespondWithJSON mengirimkan response dalam format JSON
func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
