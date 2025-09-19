package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Standard error codes for the trueblocks-dalleserver
const (
	// Client errors (400-level)
	ErrorInvalidRequest   = "INVALID_REQUEST"
	ErrorInvalidSeries    = "INVALID_SERIES"
	ErrorInvalidAddress   = "INVALID_ADDRESS"
	ErrorMissingParameter = "MISSING_PARAMETER"

	// Server errors (500-level)
	ErrorInternalServer    = "INTERNAL_SERVER_ERROR"
	ErrorFileSystem        = "FILE_SYSTEM_ERROR"
	ErrorTemplateExecution = "TEMPLATE_ERROR"

	// External service errors (502-504)
	ErrorOpenAITimeout     = "OPENAI_TIMEOUT"
	ErrorOpenAIRateLimit   = "OPENAI_RATE_LIMIT"
	ErrorOpenAIUnavailable = "OPENAI_UNAVAILABLE"
	ErrorImageDownload     = "IMAGE_DOWNLOAD_ERROR"

	// Timeout errors (408)
	ErrorTimeout           = "TIMEOUT_ERROR"
	ErrorGenerationTimeout = "GENERATION_TIMEOUT"
)

// APIError represents a structured error response
type APIError struct {
	Code      string `json:"code"`                 // Machine-readable error code
	Message   string `json:"message"`              // Human-readable message
	Details   string `json:"details,omitempty"`    // Additional context
	Timestamp int64  `json:"timestamp"`            // Unix timestamp
	RequestID string `json:"request_id,omitempty"` // For tracing
}

// APIResponse represents a structured API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// NewAPIError creates a new APIError with timestamp
func NewAPIError(code, message, details string) *APIError {
	return &APIError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().Unix(),
	}
}

// WithRequestID adds a request ID to the error
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WriteErrorResponse writes a structured error response to the HTTP response writer
func WriteErrorResponse(w http.ResponseWriter, err *APIError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	if wrapper, ok := w.(*ResponseWriterWrapper); ok {
		wrapper.statusCode = statusCode
	}
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success:   false,
		Error:     err,
		RequestID: err.RequestID,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(response)
}

// WriteSuccessResponse writes a structured success response
func WriteSuccessResponse(w http.ResponseWriter, data interface{}, requestID string) {
	w.Header().Set("Content-Type", "application/json")

	response := APIResponse{
		Success:   true,
		Data:      data,
		RequestID: requestID,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(response)
}

// Common error constructors for frequent use cases
func ErrorInvalidSeriesName(series string) *APIError {
	return NewAPIError(
		ErrorInvalidSeries,
		"Invalid series name",
		fmt.Sprintf("Series '%s' not found in available series", series),
	)
}

func ErrorInvalidAddressFormat(address string) *APIError {
	return NewAPIError(
		ErrorInvalidAddress,
		"Invalid address format",
		fmt.Sprintf("Address '%s' is not a valid Ethereum address", address),
	)
}

func ErrorMissingRequiredParameter(param string) *APIError {
	return NewAPIError(
		ErrorMissingParameter,
		"Missing required parameter",
		fmt.Sprintf("Parameter '%s' is required", param),
	)
}

func ErrorOpenAIServiceTimeout(operation string) *APIError {
	return NewAPIError(
		ErrorOpenAITimeout,
		"OpenAI service timeout",
		fmt.Sprintf("Timeout during %s operation", operation),
	)
}

func ErrorFileSystemOperation(operation, details string) *APIError {
	return NewAPIError(
		ErrorFileSystem,
		"File system operation failed",
		fmt.Sprintf("%s: %s", operation, details),
	)
}

// GenerateRequestID creates a new request ID for tracing
func GenerateRequestID() string {
	return uuid.New().String()[:8] // Short UUID for readability
}
