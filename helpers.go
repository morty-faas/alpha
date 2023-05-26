package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func prettyJson(b []byte) []byte {
	var out bytes.Buffer
	json.Indent(&out, b, "", " ")
	return out.Bytes()
}

func JSONResponse(w http.ResponseWriter, data interface{}) {
	JSONResponseWithStatusCode(w, http.StatusOK, data)
}

func JSONResponseWithStatusCode(w http.ResponseWriter, status int, data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	w.Write(prettyJson(body))
}
