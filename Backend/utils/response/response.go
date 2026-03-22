package response

import (
	"encoding/json"
	"net/http"
)

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func GeneralErrorResponse(w http.ResponseWriter, statusCode int, message string) error {
	response := map[string]string{"error": message}
	return WriteJSONResponse(w, statusCode, response)
}