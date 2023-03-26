package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/thomasgouveia/go-config"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	rootEndpoint   = "/"
	healthEndpoint = "/_/health"
)

func main() {
	cl, err := config.NewLoader(loaderOptions)
	if err != nil {
		log.Fatalf("failed to create configuration loader: %v", err)
	}

	cfg, err := cl.Load()
	if err != nil {
		log.Fatalf("failed to load config : %v", err)
	}

	level, err := log.ParseLevel(cfg.Server.LogLevel)
	if err != nil {
		level = log.InfoLevel
	}
	logger := log.New()
	logger.SetLevel(level)

	command := cfg.Process.Command
	if command == "" {
		logger.Fatalf("Unable to find a valid command in configuration. Please set ALPHA_INVOKE environment variable and restart.")
	}

	// Build the process to launch in background.
	// As currently we support only HTTP mode,
	// we assume that the process that will be launched
	// will expose an HTTP endpoint where all the incoming requests to alpha
	// will be forwarded directly. We should map the stdout / stderr
	// of the forked process to our current stream, to be able to check execution logs.
	args := strings.Split(command, " ")

	// For example here, if the command is `node /path/to/my/script.js`
	// args[0] will be node and args[1:] will be the rest of arguments.
	cmd := exec.Command(args[0], args[1:]...)
	logger.Debugf("Command : %s", cmd.String())

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Fatalf("Failed to bind command stdout to Alpha: %v", err)
	}
	cmd.Stderr = cmd.Stdout

	// Fork the process and start it
	// This will not wait for the process completion
	// so this is a non-blocking operation
	if err := cmd.Start(); err != nil {
		logger.Fatalf("Failed to start command: %v", err)
	}

	logger.Infof("Started process %s (pid=%d)", cmd.String(), cmd.Process.Pid)

	// Configure a scanner to read the downstream process logs
	sc := bufio.NewScanner(stdout)
	sc.Split(bufio.ScanLines)

	var logs []string
	// Run into a Go routine because `sc.Scan()` is blocking
	go func() {
		for sc.Scan() {
			l := sc.Text()

			// TODO: we currently write every forked process logs to the stdout.
			// It would be great if we can determine the level of the underlying logs
			// We put a tag before printing the log line to identify clearly the downstream logs
			// It will be useful later to collect logs
			fmt.Println(fmt.Sprintf("downstream: %s", l))
			logs = append(logs, l)
		}
	}()

	// If we're not able to parse correctly this URL (potentially due to a misconfiguration)
	// we should exit immediately as the server will not be able to forward requests
	// to the downstream service.
	remote, err := url.Parse(cfg.Process.DownstreamURL)
	if err != nil {
		logger.Fatalf("Failed to parse downstream URL: %v", err)
	}

	// Declare the start time of the function invocation
	// It will be initialized to time.Now before proxying the request
	var start time.Time

	// Create the reverse proxy to forward requests to our downstream process
	proxy := httputil.NewSingleHostReverseProxy(remote)

	// Configure a response interceptor to inject instrumentation metadata
	// into the payload before returning it to the caller.
	proxy.ModifyResponse = func(r *http.Response) error {
		elapsed := time.Since(start).Milliseconds()

		// As we can't predict the format of the payload returned
		// by the upstream, we use `any` here to allow json unmarshalling
		// We assume that the downstream runtime returns a JSON encodable payload
		var payload any

		by, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		// Try to parse the bytes as JSON
		// If an error occurs, we don't want to return an error
		// because it means that the server has answered another format
		if err := json.Unmarshal(by, &payload); err != nil {
			logger.Warnf("failed to parse downstream response as JSON: %v. Payload will be evaluated as text value", err)
			payload = string(by)
		}

		output := &InstrumentedResponse{
			Payload: payload,
			Process: &ProcessMetadata{
				ExecutionTimeMs: elapsed,
				Logs:            logs,
			},
		}

		// Serialize the computed response
		body, err := json.Marshal(output)
		if err != nil {
			return err
		}
		contentLength := len(body)

		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(contentLength)
		r.Header.Set("Content-Length", strconv.Itoa(contentLength))

		// Remove any information on the underlying runtime
		// to improve security
		r.Header.Del("X-Powered-By")

		// Reset all instrumentation metadata for
		// future invocations
		logs = []string{}
		start = time.Now()

		return nil
	}

	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Host = remote.Host

			start = time.Now()
			p.ServeHTTP(w, r)
		}
	}

	ctx, stop := context.WithCancel(context.Background())

	r := http.NewServeMux()
	r.HandleFunc(rootEndpoint, handler(proxy))
	r.HandleFunc(healthEndpoint, healthHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.Server.Port),
		Handler: r,
	}

	// Listen for signals to trigger graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// The go routine will be blocked until the program will receive one of the signals declared above.
	// This routine will try to exit gracefully Alpha by closing connections and
	// clean up the forked process.
	go func() {
		<-sigs

		logger.Infof("Alpha is trying to exit gracefully. Hit CTRL+C to force exiting.")

		shutdownCtx, _ := context.WithTimeout(ctx, 30*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Fatal("Graceful shutdown timed out. Forcing exit")
			}
		}()

		logger.Info("Shutting down internal HTTP server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Fatal(err)
		}

		logger.Info("Cleaning up downstream process")
		if err := cmd.Process.Kill(); err != nil {
			// We don't want to fatal here in the case we can't clean up the forked process.
			// We simply print an error with the PID, so the user will be able to clean up the process itself.
			logger.Errorf("Failed to clean up forked process with pid=%d: %v", cmd.Process.Pid, err)
		}

		stop()
	}()

	logger.Infof("Alpha server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}

	<-ctx.Done()
}
