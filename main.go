package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/polyxia-org/alpha/internal/config"
	log "github.com/sirupsen/logrus"
)

const (
	rootEndpoint   = "/"
	healthEndpoint = "/_/health"
)

func main() {
	// Loading configuration from environment
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config : %v", err)
	}

	logger := log.New()
	logger.SetLevel(config.LogLevel)

	command := config.InvokeInstruction
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

	// Run into a Go routine because `sc.Scan()` is blocking
	go func() {
		for sc.Scan() {
			m := sc.Text()

			// TODO: we currently write every forked process logs to the stdout.
			// It would be great if we can determine the level of the underlying logs
			// We put a tag before printing the log line to identify clearly the downstream logs
			// It will be useful later to collect logs
			fmt.Println(fmt.Sprintf("downstream: %s", m))
		}
	}()

	// If we're not able to parse correctly this URL (potentially due to a misconfiguration)
	// we should exit immediately as the server will not be able to forward requests
	// to the downstream service.
	remote, err := url.Parse(config.Remote)
	if err != nil {
		logger.Fatalf("Failed to parse downstream URL: %v", err)
	}

	// Configure the reverse proxy to forward requests to our downstream process
	proxy := httputil.NewSingleHostReverseProxy(remote)
	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Host = remote.Host
			p.ServeHTTP(w, r)
		}
	}

	ctx, stop := context.WithCancel(context.Background())

	r := http.NewServeMux()
	r.HandleFunc(rootEndpoint, handler(proxy))
	r.HandleFunc(healthEndpoint, healthHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.Port),
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
