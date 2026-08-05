package responces

import (
	"encoding/json"
	"net/http"
)

func ResponseJson(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

type ErrorBody struct {
	Error struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error"`
}

func ResponseError(w http.ResponseWriter, status int, code, message string) {
	var resp ErrorBody

	resp.Error.Code = code
	resp.Error.Message = message

	ResponseJson(w, status, resp)
}
