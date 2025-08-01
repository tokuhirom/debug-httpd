//go:build integration
// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

const (
	testPort      = "19876"
	testImageName = "debug-httpd:integration-test"
	containerName = "debug-httpd-integration-test"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Clean up any existing container
	cleanupContainer()

	// Build Docker image
	t.Log("Building Docker image...")
	buildCmd := exec.Command("docker", "build", "-t", testImageName, ".")
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build Docker image: %v\nOutput: %s", err, output)
	}

	// Run Docker container
	t.Log("Starting Docker container...")
	runCmd := exec.Command("docker", "run", "-d", "--name", containerName,
		"-p", fmt.Sprintf("%s:%s", testPort, testPort),
		testImageName, testPort)
	output, err = runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run Docker container: %v\nOutput: %s", err, output)
	}

	// Ensure cleanup
	defer cleanupContainer()

	// Wait for container to be ready
	t.Log("Waiting for container to be ready...")
	waitForServer(t, fmt.Sprintf("http://localhost:%s/ping", testPort), 30*time.Second)

	// Run test cases
	t.Run("PingEndpoint", func(t *testing.T) {
		testPingEndpoint(t)
	})

	t.Run("DebugEndpoint", func(t *testing.T) {
		testDebugEndpoint(t)
	})

	t.Run("LogsEndpoint", func(t *testing.T) {
		testLogsEndpoint(t)
	})

	t.Run("PortConfiguration", func(t *testing.T) {
		testPortConfiguration(t)
	})
}

func cleanupContainer() {
	// Stop container
	exec.Command("docker", "stop", containerName).Run()
	// Remove container
	exec.Command("docker", "rm", containerName).Run()
}

func waitForServer(t *testing.T, url string, timeout time.Duration) {
	start := time.Now()
	for {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if time.Since(start) > timeout {
			t.Fatalf("Server did not become ready within %v", timeout)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func testPingEndpoint(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/ping", testPort))
	if err != nil {
		t.Fatalf("Failed to access /ping: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	if body != "pong" {
		t.Errorf("Expected 'pong', got '%s'", body)
	}
}

func testDebugEndpoint(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/", testPort))
	if err != nil {
		t.Fatalf("Failed to access /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Check required fields
	requiredFields := []string{"timestamp", "request", "host", "environment_variables", "go_version"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Missing required field: %s", field)
		}
	}

	// Check host information
	if host, ok := result["host"].(map[string]interface{}); ok {
		if _, ok := host["hostname"]; !ok {
			t.Error("Missing hostname in host information")
		}
		if _, ok := host["ip_addresses"]; !ok {
			t.Error("Missing ip_addresses in host information")
		}
	} else {
		t.Error("Host field is not a map")
	}

	// Check environment variables include PORT
	if envVars, ok := result["environment_variables"].(map[string]interface{}); ok {
		if port, ok := envVars["PORT"]; ok {
			if port != testPort {
				t.Errorf("Expected PORT=%s, got %v", testPort, port)
			}
		}
	}
}

func testLogsEndpoint(t *testing.T) {
	// Make a few requests first
	http.Get(fmt.Sprintf("http://localhost:%s/test1", testPort))
	http.Get(fmt.Sprintf("http://localhost:%s/test2", testPort))
	
	// Get logs
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/logs", testPort))
	if err != nil {
		t.Fatalf("Failed to access /logs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var logs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		t.Fatalf("Failed to decode logs JSON: %v", err)
	}

	if len(logs) < 3 {
		t.Errorf("Expected at least 3 log entries, got %d", len(logs))
	}

	// Check log structure
	if len(logs) > 0 {
		requiredFields := []string{"timestamp", "method", "path", "client_address", "user_agent"}
		for _, field := range requiredFields {
			if _, ok := logs[0][field]; !ok {
				t.Errorf("Missing required field in log entry: %s", field)
			}
		}
	}

	// Check if our test requests are in the logs
	foundTest1 := false
	foundTest2 := false
	for _, log := range logs {
		if path, ok := log["path"].(string); ok {
			if path == "/test1" {
				foundTest1 = true
			}
			if path == "/test2" {
				foundTest2 = true
			}
		}
	}

	if !foundTest1 || !foundTest2 {
		t.Error("Test requests not found in logs")
	}
}

func testPortConfiguration(t *testing.T) {
	// Test with environment variable
	t.Log("Testing port configuration with environment variable...")
	
	// Clean up
	cleanupContainer()
	
	// Run with PORT environment variable
	containerNameEnv := containerName + "-env"
	testPortEnv := "29876"
	
	runCmd := exec.Command("docker", "run", "-d", "--name", containerNameEnv,
		"-e", fmt.Sprintf("PORT=%s", testPortEnv),
		"-p", fmt.Sprintf("%s:%s", testPortEnv, testPortEnv),
		testImageName)
	output, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run container with PORT env: %v\nOutput: %s", err, output)
	}
	
	// Cleanup
	defer func() {
		exec.Command("docker", "stop", containerNameEnv).Run()
		exec.Command("docker", "rm", containerNameEnv).Run()
	}()
	
	// Wait for server
	waitForServer(t, fmt.Sprintf("http://localhost:%s/ping", testPortEnv), 30*time.Second)
	
	// Test ping
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/ping", testPortEnv))
	if err != nil {
		t.Fatalf("Failed to access server on custom port %s: %v", testPortEnv, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 on custom port, got %d", resp.StatusCode)
	}
}