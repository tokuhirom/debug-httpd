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

func TestSleepHandler_Success(t *testing.T) {
	tests := []struct {
		name         string
		duration     string
		minSleepTime time.Duration
		maxSleepTime time.Duration
	}{
		{
			name:         "sleep 100ms",
			duration:     "100ms",
			minSleepTime: 90 * time.Millisecond,
			maxSleepTime: 150 * time.Millisecond,
		},
		{
			name:         "sleep 1s",
			duration:     "1s",
			minSleepTime: 900 * time.Millisecond,
			maxSleepTime: 1200 * time.Millisecond,
		},
		{
			name:         "sleep 500ms",
			duration:     "500ms",
			minSleepTime: 450 * time.Millisecond,
			maxSleepTime: 600 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/sleep?duration="+tt.duration, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(sleepHandler)

			start := time.Now()
			handler.ServeHTTP(rr, req)
			elapsed := time.Since(start)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			// Check that the handler actually slept for approximately the right duration
			if elapsed < tt.minSleepTime {
				t.Errorf("handler slept for too short: got %v, want at least %v",
					elapsed, tt.minSleepTime)
			}
			if elapsed > tt.maxSleepTime {
				t.Errorf("handler slept for too long: got %v, want at most %v",
					elapsed, tt.maxSleepTime)
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("could not parse response: %v", err)
			}

			// Check response fields
			if _, ok := response["slept_duration"]; !ok {
				t.Error("response missing slept_duration field")
			}
			if _, ok := response["actual_duration"]; !ok {
				t.Error("response missing actual_duration field")
			}
			if _, ok := response["timestamp"]; !ok {
				t.Error("response missing timestamp field")
			}
		})
	}
}

func TestSleepHandler_MissingDuration(t *testing.T) {
	req, err := http.NewRequest("GET", "/sleep", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(sleepHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("could not parse response: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("response missing error field")
	}
}

func TestSleepHandler_InvalidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"invalid format", "invalid"},
		{"negative duration", "-1s"},
		{"exceeds max", "2h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/sleep?duration="+tt.duration, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(sleepHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusBadRequest {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusBadRequest)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("could not parse response: %v", err)
			}

			if _, ok := response["error"]; !ok {
				t.Error("response missing error field")
			}
		})
	}
}

func TestStatusHandler_Success(t *testing.T) {
	tests := []struct {
		name               string
		code               string
		message            string
		expectedStatusCode int
		expectedMessage    string
	}{
		{
			name:               "status 200",
			code:               "200",
			message:            "",
			expectedStatusCode: 200,
			expectedMessage:    "OK",
		},
		{
			name:               "status 404",
			code:               "404",
			message:            "",
			expectedStatusCode: 404,
			expectedMessage:    "Not Found",
		},
		{
			name:               "status 500",
			code:               "500",
			message:            "",
			expectedStatusCode: 500,
			expectedMessage:    "Internal Server Error",
		},
		{
			name:               "status 201 with custom message",
			code:               "201",
			message:            "Custom Created Message",
			expectedStatusCode: 201,
			expectedMessage:    "Custom Created Message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/status?code=" + tt.code
			if tt.message != "" {
				url += "&message=" + tt.message
			}

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(statusHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatusCode)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("could not parse response: %v", err)
			}

			// Check response fields
			if statusCode, ok := response["status_code"].(float64); !ok || int(statusCode) != tt.expectedStatusCode {
				t.Errorf("unexpected status_code: got %v want %v", response["status_code"], tt.expectedStatusCode)
			}

			if message, ok := response["message"].(string); !ok || message != tt.expectedMessage {
				t.Errorf("unexpected message: got %v want %v", response["message"], tt.expectedMessage)
			}

			if _, ok := response["timestamp"]; !ok {
				t.Error("response missing timestamp field")
			}
		})
	}
}

func TestStatusHandler_MissingCode(t *testing.T) {
	req, err := http.NewRequest("GET", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(statusHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("could not parse response: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("response missing error field")
	}
}

func TestStatusHandler_InvalidCode(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"invalid format", "invalid"},
		{"below range", "99"},
		{"above range", "600"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/status?code="+tt.code, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(statusHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusBadRequest {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusBadRequest)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("could not parse response: %v", err)
			}

			if _, ok := response["error"]; !ok {
				t.Error("response missing error field")
			}
		})
	}
}