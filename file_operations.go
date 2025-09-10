package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
)

// FileSystemError represents file system operation errors
type FileSystemError struct {
	Operation string
	Path      string
	Err       error
	RequestID string
}

func (e *FileSystemError) Error() string {
	return fmt.Sprintf("[%s] file system error during %s on '%s': %v", e.RequestID, e.Operation, e.Path, e.Err)
}

func (e *FileSystemError) Unwrap() error {
	return e.Err
}

// IsRetryable determines if the file system error can be retried
func (e *FileSystemError) IsRetryable() bool {
	if e.Err == nil {
		return false
	}

	// Check for specific error types that might be retryable
	if errno, ok := e.Err.(syscall.Errno); ok {
		switch errno {
		case syscall.EBUSY, syscall.EAGAIN, syscall.EINTR:
			return true // Temporary resource issues
		case syscall.ENOSPC, syscall.EACCES, syscall.EPERM:
			return false // Disk full, permissions - not retryable
		}
	}

	return false
}

// RobustFileOperations provides resilient file system operations
type RobustFileOperations struct {
	retryConfig RetryConfig
}

// NewRobustFileOperations creates a new instance with default retry configuration
func NewRobustFileOperations() *RobustFileOperations {
	return &RobustFileOperations{
		retryConfig: RetryConfig{
			MaxAttempts:   3,
			BaseDelay:     100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffFactor: 2.0,
		},
	}
}

// EnsureDirectory creates directory with robust error handling
func (rfo *RobustFileOperations) EnsureDirectory(dirPath, requestID string) error {
	operation := func() error {
		if err := os.MkdirAll(dirPath, 0o750); err != nil {
			return &FileSystemError{
				Operation: "mkdir",
				Path:      dirPath,
				Err:       err,
				RequestID: requestID,
			}
		}
		return nil
	}

	return RetryWithBackoff(rfo.retryConfig, operation)
}

// validateFilePath validates a file path to prevent directory traversal attacks
func validateFilePath(filePath string) error {
	// Check for directory traversal attempts
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid file path: directory traversal detected")
	}

	// Additional validation can be added here
	return nil
}

// WriteFile writes data to file with robust error handling
func (rfo *RobustFileOperations) WriteFile(filePath string, data []byte, requestID string) error {
	operation := func() error {
		// Validate file path to prevent directory traversal
		if err := validateFilePath(filePath); err != nil {
			return &FileSystemError{
				Operation: "validate_path",
				Path:      filePath,
				Err:       err,
				RequestID: requestID,
			}
		}

		// Ensure parent directory exists
		dir := filepath.Dir(filePath)
		if err := rfo.EnsureDirectory(dir, requestID); err != nil {
			return err
		}

		// Create temporary file first for atomic write
		tempFile := filePath + ".tmp"

		file, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 - validated above
		if err != nil {
			return &FileSystemError{
				Operation: "create_temp",
				Path:      tempFile,
				Err:       err,
				RequestID: requestID,
			}
		}

		_, writeErr := file.Write(data)
		closeErr := file.Close()

		if writeErr != nil {
			_ = os.Remove(tempFile) // Clean up temp file
			return &FileSystemError{
				Operation: "write",
				Path:      tempFile,
				Err:       writeErr,
				RequestID: requestID,
			}
		}

		if closeErr != nil {
			_ = os.Remove(tempFile) // Clean up temp file
			return &FileSystemError{
				Operation: "close",
				Path:      tempFile,
				Err:       closeErr,
				RequestID: requestID,
			}
		}

		// Atomic move from temp to final location
		if err := os.Rename(tempFile, filePath); err != nil {
			_ = os.Remove(tempFile) // Clean up temp file
			return &FileSystemError{
				Operation: "rename",
				Path:      filePath,
				Err:       err,
				RequestID: requestID,
			}
		}

		return nil
	}

	return RetryWithBackoff(rfo.retryConfig, operation)
}

// ReadFile reads file with robust error handling
func (rfo *RobustFileOperations) ReadFile(filePath, requestID string) ([]byte, error) {
	var data []byte

	operation := func() error {
		// Validate file path to prevent directory traversal
		if err := validateFilePath(filePath); err != nil {
			return &FileSystemError{
				Operation: "validate_path",
				Path:      filePath,
				Err:       err,
				RequestID: requestID,
			}
		}

		file, err := os.Open(filePath) // #nosec G304 - validated above
		if err != nil {
			return &FileSystemError{
				Operation: "open",
				Path:      filePath,
				Err:       err,
				RequestID: requestID,
			}
		}
		defer file.Close()

		fileData, err := io.ReadAll(file)
		if err != nil {
			return &FileSystemError{
				Operation: "read",
				Path:      filePath,
				Err:       err,
				RequestID: requestID,
			}
		}

		data = fileData
		return nil
	}

	err := RetryWithBackoff(rfo.retryConfig, operation)
	return data, err
}

// CopyFile copies file with robust error handling
func (rfo *RobustFileOperations) CopyFile(srcPath, dstPath, requestID string) error {
	operation := func() error {
		// Read source file
		data, err := rfo.ReadFile(srcPath, requestID)
		if err != nil {
			return err
		}

		// Write to destination
		return rfo.WriteFile(dstPath, data, requestID)
	}

	return RetryWithBackoff(rfo.retryConfig, operation)
}

// RemoveFile removes file with robust error handling
func (rfo *RobustFileOperations) RemoveFile(filePath, requestID string) error {
	operation := func() error {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return &FileSystemError{
				Operation: "remove",
				Path:      filePath,
				Err:       err,
				RequestID: requestID,
			}
		}
		return nil
	}

	return RetryWithBackoff(rfo.retryConfig, operation)
}

// CheckDiskSpace checks available disk space for a given path
func (rfo *RobustFileOperations) CheckDiskSpace(path, requestID string) (*DiskSpaceInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, &FileSystemError{
			Operation: "statfs",
			Path:      path,
			Err:       err,
			RequestID: requestID,
		}
	}

	info := &DiskSpaceInfo{
		Total:     stat.Blocks * uint64(stat.Bsize),
		Available: stat.Bavail * uint64(stat.Bsize),
		Used:      (stat.Blocks - stat.Bfree) * uint64(stat.Bsize),
	}

	info.UsedPercent = float64(info.Used) / float64(info.Total) * 100

	return info, nil
}

// ValidateWriteAccess checks if we can write to a directory
func (rfo *RobustFileOperations) ValidateWriteAccess(dirPath, requestID string) error {
	testFile := filepath.Join(dirPath, ".write_test_"+requestID)

	// Try to write a test file
	err := rfo.WriteFile(testFile, []byte("test"), requestID)
	if err != nil {
		return err
	}

	// Clean up test file
	_ = rfo.RemoveFile(testFile, requestID)

	logger.Info(fmt.Sprintf("[%s] Write access validated for %s", requestID, dirPath))
	return nil
}

// DiskSpaceInfo holds disk space information
type DiskSpaceInfo struct {
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Available   uint64  `json:"available_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

// IsLowSpace checks if available disk space is below threshold
func (info *DiskSpaceInfo) IsLowSpace(thresholdPercent float64) bool {
	return info.UsedPercent > thresholdPercent
}

// Global file operations instance
var globalFileOps = NewRobustFileOperations()

// GetFileOperations returns the global file operations instance
func GetFileOperations() *RobustFileOperations {
	return globalFileOps
}
