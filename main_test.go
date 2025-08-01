package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPingHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(pingHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "pong"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "text/plain")
	}
}

func TestLogsHandler(t *testing.T) {
	// Reset logger for test
	logger = NewAccessLogger(100)

	// Add some test logs
	logger.Add(AccessLog{
		Timestamp:     time.Now().Format(time.RFC3339Nano),
		Method:        "GET",
		Path:          "/test1",
		ClientAddress: "127.0.0.1",
		ClientPort:    8080,
		UserAgent:     "test-agent",
		Referer:       "",
		Host:          "localhost",
	})
	logger.Add(AccessLog{
		Timestamp:     time.Now().Format(time.RFC3339Nano),
		Method:        "POST",
		Path:          "/test2",
		ClientAddress: "127.0.0.1",
		ClientPort:    8081,
		UserAgent:     "test-agent-2",
		Referer:       "http://example.com",
		Host:          "localhost",
	})

	req, err := http.NewRequest("GET", "/logs", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(logsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var logs []AccessLog
	if err := json.Unmarshal(rr.Body.Bytes(), &logs); err != nil {
		t.Errorf("could not parse response: %v", err)
	}

	// We should have at least 3 logs (2 we added + 1 from the /logs request itself)
	if len(logs) < 3 {
		t.Errorf("expected at least 3 logs, got %d", len(logs))
	}

	// Check if our test logs are present
	foundTest1 := false
	foundTest2 := false
	for _, log := range logs {
		if log.Path == "/test1" && log.Method == "GET" {
			foundTest1 = true
		}
		if log.Path == "/test2" && log.Method == "POST" {
			foundTest2 = true
		}
	}

	if !foundTest1 {
		t.Error("did not find test1 log entry")
	}
	if !foundTest2 {
		t.Error("did not find test2 log entry")
	}
}

func TestDebugHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "test-user-agent")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(debugHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("could not parse response: %v", err)
	}

	// Check required fields exist
	requiredFields := []string{"timestamp", "request", "host", "environment_variables", "go_version"}
	for _, field := range requiredFields {
		if _, ok := response[field]; !ok {
			t.Errorf("response missing required field: %s", field)
		}
	}

	// Check request details
	if request, ok := response["request"].(map[string]interface{}); ok {
		if path, ok := request["path"].(string); !ok || path != "/" {
			t.Errorf("unexpected request path: %v", path)
		}
	} else {
		t.Error("request field is not a map")
	}
}

func TestAccessLogger(t *testing.T) {
	logger := NewAccessLogger(3)

	// Add 5 logs to test the size limit
	for i := 0; i < 5; i++ {
		logger.Add(AccessLog{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			Method:    "GET",
			Path:      "/test",
		})
	}

	logs := logger.GetLogs()
	if len(logs) != 3 {
		t.Errorf("expected 3 logs (size limit), got %d", len(logs))
	}
}

func TestGetIPAddresses(t *testing.T) {
	ips := getIPAddresses()
	if len(ips) == 0 {
		t.Error("expected at least one IP address")
	}

	// Check that at least loopback is present
	hasLoopback := false
	for _, ip := range ips {
		if ip == "127.0.0.1" || ip == "::1" {
			hasLoopback = true
			break
		}
	}
	if !hasLoopback {
		t.Error("expected to find loopback address")
	}
}