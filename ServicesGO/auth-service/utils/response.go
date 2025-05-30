package utils

import (
	"encoding/json"
	"net/http"
)

// APIError — структура для ошибок
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespondJSON отправляет обычный JSON-ответ
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// RespondError отправляет JSON-ошибку по коду и сообщению
func RespondError(w http.ResponseWriter, status int, code, message string) {
	RespondJSON(w, status, APIError{Code: code, Message: message})
}
