package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TrueBlocks/trueblocks-art/packages/ai"
)

type enhanceTransport func(*http.Request) (*http.Response, error)

func (f enhanceTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func enhanceResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func enhancementClient(t *testing.T, transport enhanceTransport) *OpenAIClient {
	t.Helper()
	t.Setenv("TRUEBLOCKS_DATA_DIR", t.TempDir())
	c := NewOpenAIClient("fake")
	c.httpClient = &http.Client{Transport: transport}
	c.endpoint = "https://fixture.invalid/enhance"
	c.circuitBreaker = NewCircuitBreaker(2, time.Minute)
	c.metrics = NewMetricsCollector()
	c.retryConfig.BaseDelay = time.Millisecond
	c.retryConfig.MaxDelay = time.Millisecond
	return c
}

func TestSharedEnhancementSuccessAndAccounting(t *testing.T) {
	calls := 0
	c := enhancementClient(t, func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.String() != "https://fixture.invalid/enhance" || r.Header.Get("X-Request-ID") != "caller" {
			t.Fatal("endpoint or correlation changed")
		}
		var body struct {
			Model       string
			Seed        int
			Temperature float64
			MaxTokens   *int `json:"max_tokens"`
			Messages    []struct{ Role, Content string }
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Model != "gpt-4" || body.Seed != 1337 || body.Temperature != 0.2 || body.MaxTokens != nil || len(body.Messages) != 1 || body.Messages[0].Role != "system" || body.Messages[0].Content != "prompt" {
			t.Fatalf("%+v", body)
		}
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > enhanceDeadline {
			t.Fatal("per-attempt timeout missing")
		}
		return enhanceResponse(200, `{"choices":[{"message":{"content":"enhanced"}}],"usage":{"prompt_tokens":1000,"completion_tokens":100}}`), nil
	})
	got, err := c.EnhancePromptWithResilience("prompt", "ignored author", "caller")
	if err != nil || got != "enhanced" || calls != 1 {
		t.Fatalf("%s calls=%d err=%v", got, calls, err)
	}
	m := c.metrics.GetMetrics()
	if m.OpenAIRequests != 1 || m.OpenAIErrors != 0 || m.TotalRetries != 0 {
		t.Fatalf("%+v", m)
	}
	data, err := os.ReadFile(ai.DefaultLedgerPath())
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil || len(rows) != 2 {
		t.Fatalf("%v %v", rows, err)
	}
	values := map[string]string{}
	for i, key := range rows[0] {
		values[key] = rows[1][i]
	}
	for key, want := range map[string]string{"tool": "dalleserver", "compose_model": "gpt-4", "compose_in": "1000", "compose_out": "100", "usd": "0.0360"} {
		if values[key] != want {
			t.Fatalf("%s=%s want=%s", key, values[key], want)
		}
	}
	if m := c.GetCircuitBreakerMetrics(); m.TotalSuccesses != 1 || m.TotalFailures != 0 {
		t.Fatalf("%+v", m)
	}
}

func TestSharedEnhancementRetriesAndFallback(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		status, failures, wantCalls int
		body                        string
	}{
		{"rate recovery", 429, 2, 3, `{"error":{"code":"rate_limit","message":"slow down"}}`},
		{"server recovery", 503, 1, 2, "unavailable"},
		{"exhaustion", 504, 9, 3, "gateway timeout"},
		{"permanent", 401, 9, 1, `{"error":{"code":"invalid_api_key","message":"bad key"}}`},
		{"quota", 429, 9, 1, `{"error":{"code":"insufficient_quota","message":"no balance"}}`},
		{"network", 0, 1, 2, ""},
		{"malformed", 200, 9, 1, "not JSON"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			c := enhancementClient(t, func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Header.Get("X-Request-ID") != "caller" {
					t.Fatal("retry lost request ID")
				}
				body, err := io.ReadAll(r.Body)
				if err != nil || len(body) == 0 {
					t.Fatal("retry body missing")
				}
				if calls <= tc.failures {
					if tc.status == 0 {
						return nil, &net.OpError{Op: "read", Net: "tcp", Err: io.ErrUnexpectedEOF}
					}
					return enhanceResponse(tc.status, tc.body), nil
				}
				return enhanceResponse(200, `{"choices":[{"message":{"content":"enhanced"}}],"usage":{"prompt_tokens":10,"completion_tokens":1}}`), nil
			})
			got, err := c.EnhancePromptWithResilience("prompt", "", "caller")
			want := "prompt"
			if tc.failures < tc.wantCalls {
				want = "enhanced"
			}
			if err != nil || got != want || calls != tc.wantCalls {
				t.Fatalf("got=%s calls=%d err=%v", got, calls, err)
			}
			m := c.metrics.GetMetrics()
			failures := min(tc.failures, tc.wantCalls)
			if m.OpenAIRequests != int64(calls) || m.OpenAIErrors != int64(failures) || m.TotalRetries != int64(calls-1) || m.TotalErrors != int64(failures) {
				t.Fatalf("%+v", m)
			}
			if tc.status == 504 && m.OpenAITimeouts != 3 {
				t.Fatalf("timeouts=%d", m.OpenAITimeouts)
			}
			if want == "prompt" {
				if _, err := os.Stat(ai.DefaultLedgerPath()); !os.IsNotExist(err) {
					t.Fatal("failed call accounted")
				}
				if m := c.GetCircuitBreakerMetrics(); m.TotalFailures != 1 {
					t.Fatalf("breaker counted attempts: %+v", m)
				}
			} else {
				data, err := os.ReadFile(ai.DefaultLedgerPath())
				if err != nil || strings.Count(string(data), "\n") != 2 {
					t.Fatalf("duplicate/missing accounting: %s %v", data, err)
				}
			}
		})
	}
}

func TestSharedEnhancementCircuitAndEmptyFallback(t *testing.T) {
	t.Run("open breaker", func(t *testing.T) {
		calls := 0
		c := enhancementClient(t, func(*http.Request) (*http.Response, error) {
			calls++
			return enhanceResponse(401, "bad key"), nil
		})
		for range 3 {
			got, err := c.EnhancePromptWithResilience("prompt", "", "id")
			if err != nil || got != "prompt" {
				t.Fatalf("%s %v", got, err)
			}
		}
		if calls != 2 || c.GetCircuitBreakerMetrics().State != CircuitOpen || c.metrics.GetMetrics().OpenAIRequests != 2 {
			t.Fatal("open breaker did not stop API calls")
		}
	})
	for _, body := range []string{`{"choices":[]}`, `{"choices":[{"message":{"content":""}}],"usage":{"prompt_tokens":1,"completion_tokens":0}}`} {
		t.Run(body, func(t *testing.T) {
			c := enhancementClient(t, func(*http.Request) (*http.Response, error) { return enhanceResponse(200, body), nil })
			got, err := c.EnhancePromptWithResilience("prompt", "", "id")
			if err != nil || got != "prompt" {
				t.Fatalf("%s %v", got, err)
			}
			m := c.metrics.GetMetrics()
			if m.OpenAIRequests != 1 || m.OpenAIErrors != 0 || m.TotalRetries != 0 {
				t.Fatalf("%+v", m)
			}
		})
	}
}

func TestSharedEnhancementCancellationAndAttemptDeadline(t *testing.T) {
	t.Run("cancel before call", func(t *testing.T) {
		c := enhancementClient(t, func(*http.Request) (*http.Response, error) {
			t.Fatal("cancelled call sent request")
			return nil, context.Canceled
		})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		got, err := c.EnhancePromptWithContext(ctx, "prompt", "", "id")
		if err != nil || got != "prompt" || c.metrics.GetMetrics().OpenAIRequests != 0 {
			t.Fatalf("%s %v", got, err)
		}
	})
	t.Run("cancel during backoff", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		c := enhancementClient(t, func(*http.Request) (*http.Response, error) {
			calls++
			time.AfterFunc(10*time.Millisecond, cancel)
			return enhanceResponse(503, "unavailable"), nil
		})
		c.retryConfig.BaseDelay = time.Second
		c.retryConfig.MaxDelay = time.Second
		got, err := c.EnhancePromptWithContext(ctx, "prompt", "", "id")
		if err != nil || got != "prompt" || calls != 1 || c.metrics.GetMetrics().TotalRetries != 0 {
			t.Fatalf("%s calls=%d err=%v", got, calls, err)
		}
	})
	t.Run("per attempt", func(t *testing.T) {
		calls := 0
		c := enhancementClient(t, func(r *http.Request) (*http.Response, error) {
			calls++
			<-r.Context().Done()
			return nil, r.Context().Err()
		})
		c.timeout = 5 * time.Millisecond
		c.retryConfig.MaxDelay = time.Second
		got, err := c.EnhancePromptWithResilience("prompt", "", "id")
		if err != nil || got != "prompt" || calls != 3 {
			t.Fatalf("%s calls=%d err=%v", got, calls, err)
		}
		if m := c.metrics.GetMetrics(); m.OpenAITimeouts != 3 || m.OpenAIRequests != 3 {
			t.Fatalf("%+v", m)
		}
	})
}

func TestSharedEnhancementWrappedErrorMetrics(t *testing.T) {
	c := enhancementClient(t, nil)
	err := fmt.Errorf("outer: %w", &ai.APIError{StatusCode: 504, Code: "upstream_timeout", RequestID: "provider", Err: context.DeadlineExceeded})
	c.recordAttempt(2, err, "caller")
	m := c.metrics.GetMetrics()
	if m.ErrorsByCode["upstream_timeout"] != 1 || m.OpenAITimeouts != 1 || m.TotalRetries != 1 || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("%+v", m)
	}
}
