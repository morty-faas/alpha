package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

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

func prettyJson(b []byte) []byte {
	var out bytes.Buffer
	json.Indent(&out, b, "", " ")
	return out.Bytes()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Currently we don't have any health probes implemented.
	// We can return by default an HTTP 200 response.
	// But in a future version, we will need to define how our agent
	// can be marked as ready by the system to be able to trigger functions.
	JSONResponse(w, &HealthResponse{
		Status: "UP",
	})
}

type (
	HealthResponse struct {
		Status string `json:"status"`
	}
)

func main() {

	logger := log.New()

	invokeInstruction := os.Getenv("ALPHA_INVOKE")

	if invokeInstruction == "" {
		logger.Errorf("No invoke instruction provided, please set ALPHA_INVOKE environment variable.")
		os.Exit(1)
	}

	commandArgs := strings.Split(invokeInstruction, " ")
	cmd := exec.Command(commandArgs[0], commandArgs[1:]...)
	logger.Infof("invoke : %s", cmd.String())
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)

	go func() {
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
	}()

	remote, err := url.Parse("http://localhost:3000")
	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Host = remote.Host
			p.ServeHTTP(w, r)
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	http.HandleFunc("/", handler(proxy))
	http.HandleFunc("/_/health", healthHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}

}
