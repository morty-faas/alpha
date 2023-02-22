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
	body, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	w.Write(prettyJson(body))
}
