package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEnhancedHealthEndpoint(t *testing.T) {
	app := NewApp()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.handleHealth)

	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Debug: Print the actual response
	t.Logf("Response body: %s", rr.Body.String())

	// Parse JSON response - it's wrapped in the APIResponse structure
	var response APIResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse API response: %v", err)
	}

	// Extract the health check data from the response
	healthData, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data is not a map")
	}

	// Check basic fields
	if status, exists := healthData["status"]; !exists || status == "" {
		t.Error("Health check status is missing or empty")
	}

	if components, exists := healthData["components"]; !exists {
		t.Error("Health check components are missing")
	} else {
		componentsMap, ok := components.(map[string]interface{})
		if !ok {
			t.Error("Components is not a map")
		} else if len(componentsMap) == 0 {
			t.Error("Health check components are empty")
		} else {
			// Check that expected components are present
			expectedComponents := []string{"filesystem", "memory", "disk_space"}
			for _, component := range expectedComponents {
				if _, exists := componentsMap[component]; !exists {
					t.Errorf("Expected component %s not found in health check", component)
				}
			}
		}
	}

	// Check system health
	if system, exists := healthData["system"]; exists {
		systemMap, ok := system.(map[string]interface{})
		if ok {
			if goroutines, exists := systemMap["goroutines"]; exists {
				if g, ok := goroutines.(float64); !ok || g <= 0 {
					t.Error("System goroutines count should be positive")
				}
			}
		}
	}
}

func TestEnhancedMetricsEndpoint(t *testing.T) {
	app := NewApp()

	t.Run("JSON format", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/metrics?format=json", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.handleMetrics)

		handler.ServeHTTP(rr, req)

		// Check status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check content type
		expected := "application/json"
		if ct := rr.Header().Get("Content-Type"); ct != expected {
			t.Errorf("handler returned wrong content type: got %v want %v", ct, expected)
		}

		// Debug: Print the actual response
		t.Logf("Metrics response body: %s", rr.Body.String())

		// Parse JSON response - it's wrapped in the APIResponse structure
		var response APIResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse API response: %v", err)
		}

		// Extract the metrics data from the response
		metricsData, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Response data is not a map")
		}

		// Verify basic structure exists
		if errorsByCode, exists := metricsData["errors_by_code"]; !exists {
			t.Error("ErrorsByCode should be present")
		} else if errorsByCode == nil {
			t.Error("ErrorsByCode should not be nil")
		}

		if errorsByEndpoint, exists := metricsData["errors_by_endpoint"]; !exists {
			t.Error("ErrorsByEndpoint should be present")
		} else if errorsByEndpoint == nil {
			t.Error("ErrorsByEndpoint should not be nil")
		}
	})

	t.Run("Prometheus format", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/metrics", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.handleMetrics)

		handler.ServeHTTP(rr, req)

		// Check status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check content type
		expected := "text/plain; version=0.0.4"
		if ct := rr.Header().Get("Content-Type"); ct != expected {
			t.Errorf("handler returned wrong content type: got %v want %v", ct, expected)
		}

		// Check that response contains expected metrics
		body := rr.Body.String()
		expectedMetrics := []string{
			"dalleserver_errors_total",
			"dalleserver_retries_total",
			"dalleserver_openai_requests_total",
			"dalleserver_up",
		}

		for _, metric := range expectedMetrics {
			if !strings.Contains(body, metric) {
				t.Errorf("Expected metric %s not found in response", metric)
			}
		}
	})
}

func TestReadinessEndpoint(t *testing.T) {
	app := NewApp()

	req, err := http.NewRequest("GET", "/health?check=readiness", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.handleHealth)

	handler.ServeHTTP(rr, req)

	// Check status code - should be OK for healthy service
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that response contains status
	body := rr.Body.String()
	if !strings.Contains(body, "ready") {
		t.Error("Response should contain 'ready' status")
	}
}

func TestLivenessEndpoint(t *testing.T) {
	app := NewApp()

	req, err := http.NewRequest("GET", "/health?check=liveness", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.handleHealth)

	handler.ServeHTTP(rr, req)

	// Check status code - should always be OK if server is running
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that response contains status
	body := rr.Body.String()
	if !strings.Contains(body, "alive") {
		t.Error("Response should contain 'alive' status")
	}
}

func TestMetricsCollection(t *testing.T) {
	collector := NewMetricsCollector()

	// Test error recording
	collector.RecordError("TEST_ERROR", "test_endpoint", "test-request-123")

	metrics := collector.GetMetrics()
	if metrics.TotalErrors != 1 {
		t.Errorf("Expected 1 total error, got %d", metrics.TotalErrors)
	}

	if metrics.ErrorsByCode["TEST_ERROR"] != 1 {
		t.Errorf("Expected 1 TEST_ERROR, got %d", metrics.ErrorsByCode["TEST_ERROR"])
	}

	if metrics.ErrorsByEndpoint["test_endpoint"] != 1 {
		t.Errorf("Expected 1 error for test_endpoint, got %d", metrics.ErrorsByEndpoint["test_endpoint"])
	}

	// Test retry recording
	collector.RecordRetry("test_operation", "test-request-456")

	metrics = collector.GetMetrics()
	if metrics.TotalRetries != 1 {
		t.Errorf("Expected 1 total retry, got %d", metrics.TotalRetries)
	}

	if metrics.RetriesByOperation["test_operation"] != 1 {
		t.Errorf("Expected 1 retry for test_operation, got %d", metrics.RetriesByOperation["test_operation"])
	}

	// Test response time recording
	collector.RecordResponseTime(150, "test-request-789")

	metrics = collector.GetMetrics()
	if metrics.ResponseTimes.Count != 1 {
		t.Errorf("Expected 1 response time record, got %d", metrics.ResponseTimes.Count)
	}

	if metrics.ResponseTimes.Avg != 150.0 {
		t.Errorf("Expected average response time 150.0, got %.2f", metrics.ResponseTimes.Avg)
	}

	// Test OpenAI request recording
	collector.RecordOpenAIRequest(true, false, "test-request-openai")

	metrics = collector.GetMetrics()
	if metrics.OpenAIRequests != 1 {
		t.Errorf("Expected 1 OpenAI request, got %d", metrics.OpenAIRequests)
	}

	if metrics.OpenAIErrors != 0 {
		t.Errorf("Expected 0 OpenAI errors, got %d", metrics.OpenAIErrors)
	}

	// Test OpenAI error recording
	collector.RecordOpenAIRequest(false, true, "test-request-openai-error")

	metrics = collector.GetMetrics()
	if metrics.OpenAIRequests != 2 {
		t.Errorf("Expected 2 OpenAI requests, got %d", metrics.OpenAIRequests)
	}

	if metrics.OpenAIErrors != 1 {
		t.Errorf("Expected 1 OpenAI error, got %d", metrics.OpenAIErrors)
	}

	if metrics.OpenAITimeouts != 1 {
		t.Errorf("Expected 1 OpenAI timeout, got %d", metrics.OpenAITimeouts)
	}
}

func TestHealthChecker(t *testing.T) {
	hc := NewHealthChecker()

	// Test health check without circuit breaker
	health := hc.CheckHealth("test-request-health")

	if health.Status == "" {
		t.Error("Health status should not be empty")
	}

	if len(health.Components) == 0 {
		t.Error("Health components should not be empty")
	}

	// Test that filesystem component is present
	if _, exists := health.Components["filesystem"]; !exists {
		t.Error("Filesystem component should be present")
	}

	// Test system health
	if health.System.Goroutines <= 0 {
		t.Error("System goroutines should be positive")
	}

	if health.System.GOMAXPROCS <= 0 {
		t.Error("GOMAXPROCS should be positive")
	}

	// Test with circuit breaker
	cb := NewCircuitBreaker(3, 10*time.Second)
	hc.SetCircuitBreaker(cb)

	healthWithCB := hc.CheckHealth("test-request-health-cb")

	// Should now include OpenAI component
	if _, exists := healthWithCB.Components["openai"]; !exists {
		t.Error("OpenAI component should be present when circuit breaker is set")
	}
}

func TestPrometheusMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	// Record some test data
	collector.RecordError("TEST_ERROR", "test_endpoint", "test-123")
	collector.RecordRetry("test_operation", "test-456")
	collector.RecordResponseTime(200, "test-789")

	prometheus := collector.PrometheusMetrics()

	// Check for expected metrics
	expectedMetrics := []string{
		"dalleserver_errors_total 1",
		"dalleserver_retries_total 1",
		"dalleserver_up 1",
	}

	for _, expected := range expectedMetrics {
		if !strings.Contains(prometheus, expected) {
			t.Errorf("Expected metric '%s' not found in Prometheus output", expected)
		}
	}
}
