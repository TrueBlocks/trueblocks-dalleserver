package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v6"
)

// Test error paths on the /dalle/ handler surface proper 400s with expected messages.
func TestHandleDalleErrors(t *testing.T) {
	dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"empty"}})
	app := NewApp()

	cases := []struct {
		name       string
		path       string
		wantStatus int
		wantCode   string // New: expected error code
		wantSubstr string // For backward compatibility, check message content
	}{
		{"invalid series", "/dalle/notaseries/0xf503017d7baf7fbc0fff7492b751025c6a78179b", http.StatusBadRequest, ErrorInvalidSeries, "Invalid series name"},
		{"malformed address", "/dalle/empty/0xdeadbeef", http.StatusBadRequest, ErrorInvalidAddress, "Invalid address format"},
		{"missing address", "/dalle/empty/", http.StatusBadRequest, ErrorMissingParameter, "address"},
		{"missing both", "/dalle/", http.StatusBadRequest, ErrorInvalidRequest, "Invalid request path"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, c.path, nil)
			w := httptest.NewRecorder()
			app.handleDalleDress(w, r)
			res := w.Result()
			if res.StatusCode != c.wantStatus {
				t.Fatalf("%s: expected status %d got %d", c.name, c.wantStatus, res.StatusCode)
			}

			bodyBytes, _ := io.ReadAll(res.Body)

			// Parse the structured error response
			var apiResp APIResponse
			if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
				t.Fatalf("%s: failed to parse JSON response: %v", c.name, err)
			}

			// Check that it's an error response
			if apiResp.Success {
				t.Fatalf("%s: expected error response, got success", c.name)
			}

			if apiResp.Error == nil {
				t.Fatalf("%s: expected error object in response", c.name)
			}

			// Check error code
			if apiResp.Error.Code != c.wantCode {
				t.Fatalf("%s: expected error code %q, got %q", c.name, c.wantCode, apiResp.Error.Code)
			}

			// Check that expected substring is in message or details
			if c.wantSubstr != "" {
				errorText := apiResp.Error.Message + " " + apiResp.Error.Details
				if !strings.Contains(errorText, c.wantSubstr) {
					t.Fatalf("%s: expected error text to contain %q; got message=%q details=%q",
						c.name, c.wantSubstr, apiResp.Error.Message, apiResp.Error.Details)
				}
			}

			// Check that request ID is present
			if apiResp.RequestID == "" {
				t.Fatalf("%s: expected request ID in response", c.name)
			}
		})
	}
}
