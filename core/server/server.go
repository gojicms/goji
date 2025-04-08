package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gojicms/goji/core/utils"
)

func WriteError(w http.ResponseWriter, status int, message string, data ...any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	WriteJson(w, utils.Object{
		"error": fmt.Sprintf(message, data...),
	})
}

func WriteUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	WriteJson(w, utils.Object{
		"error": "unauthorized",
	})
}

func WriteRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusFound)
}

func WriteJson(w http.ResponseWriter, data interface{}) {
	jsonResponse, err := json.Marshal(data)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonResponse)
}
