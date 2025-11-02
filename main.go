package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// AccessLog represents a single access log entry
type AccessLog struct {
	Timestamp     string `json:"timestamp"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	ClientAddress string `json:"client_address"`
	ClientPort    int    `json:"client_port"`
	UserAgent     string `json:"user_agent"`
	Referer       string `json:"referer"`
	Host          string `json:"host"`
}

// AccessLogger manages access logs with thread safety
type AccessLogger struct {
	mu   sync.RWMutex
	logs []AccessLog
	size int
}

// NewAccessLogger creates a new access logger with specified size
func NewAccessLogger(size int) *AccessLogger {
	return &AccessLogger{
		logs: make([]AccessLog, 0, size),
		size: size,
	}
}

// Add adds a new log entry
func (al *AccessLogger) Add(log AccessLog) {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.logs = append(al.logs, log)
	if len(al.logs) > al.size {
		al.logs = al.logs[len(al.logs)-al.size:]
	}
}

// GetLogs returns a copy of all logs
func (al *AccessLogger) GetLogs() []AccessLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	result := make([]AccessLog, len(al.logs))
	copy(result, al.logs)
	return result
}

var logger = NewAccessLogger(100)

// logAccess logs the HTTP request
func logAccess(r *http.Request) {
	host, portStr, _ := net.SplitHostPort(r.RemoteAddr)
	port, _ := strconv.Atoi(portStr)

	log := AccessLog{
		Timestamp:     time.Now().Format(time.RFC3339Nano),
		Method:        r.Method,
		Path:          r.URL.Path,
		ClientAddress: host,
		ClientPort:    port,
		UserAgent:     r.Header.Get("User-Agent"),
		Referer:       r.Header.Get("Referer"),
		Host:          r.Header.Get("Host"),
	}

	logger.Add(log)

	// Also log to stdout
	fmt.Printf("[%s] %s %s from %s\n", log.Timestamp, log.Method, log.Path, r.RemoteAddr)
}

// getIPAddresses returns all IP addresses of the host
func getIPAddresses() []string {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			ips = append(ips, ipNet.IP.String())
		}
	}

	return ips
}

// pingHandler handles /ping requests
func pingHandler(w http.ResponseWriter, r *http.Request) {
	logAccess(r)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "pong")
}

// logsHandler handles /logs requests
func logsHandler(w http.ResponseWriter, r *http.Request) {
	logAccess(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	logs := logger.GetLogs()
	json.NewEncoder(w).Encode(logs)
}

// debugHandler handles all other requests with debug information
func debugHandler(w http.ResponseWriter, r *http.Request) {
	logAccess(r)

	// Collect environment variables
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		for i := 0; i < len(env); i++ {
			if env[i] == '=' {
				envVars[env[:i]] = env[i+1:]
				break
			}
		}
	}

	// Get host information
	hostname, _ := os.Hostname()

	// Prepare response
	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"request": map[string]interface{}{
			"path":    r.URL.Path,
			"headers": r.Header,
			"client_address": func() string {
				host, _, _ := net.SplitHostPort(r.RemoteAddr)
				return host
			}(),
			"client_port": func() int {
				_, portStr, _ := net.SplitHostPort(r.RemoteAddr)
				port, _ := strconv.Atoi(portStr)
				return port
			}(),
		},
		"host": map[string]interface{}{
			"hostname":     hostname,
			"fqdn":         hostname, // In Go, we'd need more complex logic for true FQDN
			"ip_addresses": getIPAddresses(),
		},
		"environment_variables": envVars,
		"go_version":            runtime.Version(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Parse command line arguments
	var port int
	flag.IntVar(&port, "port", 0, "Port to listen on")
	flag.Parse()

	// If port not specified via flag, check environment variable
	if port == 0 {
		if envPort := os.Getenv("PORT"); envPort != "" {
			var err error
			port, err = strconv.Atoi(envPort)
			if err != nil {
				port = 9876
			}
		} else {
			// Check if port is provided as argument
			if flag.NArg() > 0 {
				var err error
				port, err = strconv.Atoi(flag.Arg(0))
				if err != nil {
					port = 9876
				}
			} else {
				port = 9876
			}
		}
	}

	// Set up routes
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/logs", logsHandler)
	http.HandleFunc("/", debugHandler)

	// signal handling for SIGHUP
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)
	go func() {
		for sig := range sigCh {
			// just logging... and continue running
			log.Printf("Received signal: %s, continue running", sig)
		}
	}()

	// Start server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Debug HTTP server starting on port %d", port)
	log.Printf("Access at http://localhost:%d", port)
	log.Println("Press Ctrl-C to stop")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
