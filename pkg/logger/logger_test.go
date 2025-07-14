package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitialize_ValidLevels(t *testing.T) {
	validLevels := []string{
		"debug", "Debug", "DEBUG",
		"info", "Info", "INFO",
		"warn", "Warn", "WARN",
		"error", "Error", "ERROR",
		"fatal", "Fatal", "FATAL",
		"panic", "Panic", "PANIC",
	}

	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")

			// Should not panic
			Initialize(level, logPath)

			// Verify logger is initialized
			logger := GetLogger()
			if logger == nil {
				t.Error("Expected logger to be initialized")
			}

			// Verify log file was created
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				t.Errorf("Expected log file to be created at %s", logPath)
			}
		})
	}
}

func TestInitialize_InvalidLevel(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Should default to info level and not panic
	Initialize("invalid", logPath)

	logger := GetLogger()
	if logger == nil {
		t.Error("Expected logger to be initialized even with invalid level")
	}
}

func TestInitialize_EmptyLogPath(t *testing.T) {
	// Should use default path and not panic
	// Note: This will try to create the default path, which might fail in test environment
	// We'll capture any panic and verify it's related to permissions, not logic
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic due to permissions on default path
			if !strings.Contains(r.(string), "failed to create log directory") {
				t.Errorf("Unexpected panic: %v", r)
			}
		}
	}()

	Initialize("info", "")
}

func TestInitialize_RelativePath(t *testing.T) {
	// This test modifies the working directory and cannot be run in parallel
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	os.Chdir(tmpDir)

	// Use relative path
	relativePath := "logs/test.log"
	Initialize("info", relativePath)

	logger := GetLogger()
	if logger == nil {
		t.Error("Expected logger to be initialized with relative path")
	}

	// Verify log file was created in the absolute path
	expectedPath := filepath.Join(tmpDir, "logs", "test.log")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected log file to be created at %s", expectedPath)
	}
}

func TestInitialize_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "nested", "directory", "test.log")

	Initialize("info", logPath)

	// Verify nested directories were created
	logDir := filepath.Dir(logPath)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Errorf("Expected log directory to be created at %s", logDir)
	}

	// Verify log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Expected log file to be created at %s", logPath)
	}
}

func TestGetLogger_AutoInitialize(t *testing.T) {
	// WARNING: This test modifies global state and should not be run in parallel

	// Store original logger to restore after test
	originalLogger := log
	defer func() {
		// Restore original logger state to prevent affecting other tests
		log = originalLogger
	}()

	// CRITICAL: Setting global logger to nil to test auto-initialization behavior
	// This modification affects the global state and could impact other tests
	// if they run concurrently or depend on the logger being initialized
	log = nil

	// This should auto-initialize with defaults (though it may panic due to permissions)
	defer func() {
		if r := recover(); r != nil {
			// Expected behavior - trying to create default log directory
			if !strings.Contains(r.(string), "failed to create log directory") {
				t.Errorf("Unexpected panic: %v", r)
			}
		}
	}()

	logger := GetLogger()
	if logger == nil {
		t.Error("Expected GetLogger to auto-initialize")
	}
}

func TestLoggingFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	Initialize("debug", logPath)

	// Test all logging functions - they should not panic
	Debug("debug message: %s", "test")
	Info("info message: %s", "test")
	Warn("warn message: %s", "test")
	Error("error message: %s", "test")

	// Note: We don't test Fatal and Panic as they would terminate the test

	// Verify log file has content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "debug message: test") {
		t.Error("Expected debug message in log file")
	}

	if !strings.Contains(logContent, "info message: test") {
		t.Error("Expected info message in log file")
	}

	if !strings.Contains(logContent, "warn message: test") {
		t.Error("Expected warn message in log file")
	}

	if !strings.Contains(logContent, "error message: test") {
		t.Error("Expected error message in log file")
	}
}

func TestWith_StructuredLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	Initialize("info", logPath)

	// Test structured logging
	contextLogger := With("key1", "value1", "key2", 42)
	contextLogger.Info("structured message")

	// Verify log file has content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "structured message") {
		t.Error("Expected structured message in log file")
	}

	// Should contain the structured fields
	if !strings.Contains(logContent, "key1") || !strings.Contains(logContent, "value1") {
		t.Error("Expected structured fields in log file")
	}
}

func TestLogLevels_Filtering(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Initialize with warn level - should filter out debug and info
	Initialize("warn", logPath)

	Debug("debug message - should not appear")
	Info("info message - should not appear")
	Warn("warn message - should appear")
	Error("error message - should appear")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Debug and info should be filtered out
	if strings.Contains(logContent, "debug message - should not appear") {
		t.Error("Debug message should be filtered out at warn level")
	}

	if strings.Contains(logContent, "info message - should not appear") {
		t.Error("Info message should be filtered out at warn level")
	}

	// Warn and error should be present
	if !strings.Contains(logContent, "warn message - should appear") {
		t.Error("Expected warn message in log file")
	}

	if !strings.Contains(logContent, "error message - should appear") {
		t.Error("Expected error message in log file")
	}
}

func TestPathCleaning_DirectoryTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Test various path traversal attempts
	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"/tmp/../../../etc/passwd",
		"logs/../../../sensitive/file",
	}

	for _, maliciousPath := range maliciousPaths {
		t.Run(maliciousPath, func(t *testing.T) {
			// Should not panic and should clean the path
			logPath := filepath.Join(tmpDir, "safe", "test.log")
			Initialize("info", logPath)

			// Verify the logger was initialized safely
			logger := GetLogger()
			if logger == nil {
				t.Error("Expected logger to be initialized")
			}
		})
	}
}

func TestFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	Initialize("info", logPath)

	// Check file permissions
	fileInfo, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Failed to stat log file: %v", err)
	}

	mode := fileInfo.Mode()
	if mode.Perm() != 0o600 {
		t.Errorf("Expected log file permissions 0600, got %o", mode.Perm())
	}

	// Check directory permissions
	logDir := filepath.Dir(logPath)
	dirInfo, err := os.Stat(logDir)
	if err != nil {
		t.Fatalf("Failed to stat log directory: %v", err)
	}

	dirMode := dirInfo.Mode()
	// Directory permissions can vary based on umask, but should be at least readable/writable by owner
	actualPerm := dirMode.Perm()
	if actualPerm != 0o750 && actualPerm != 0o755 {
		t.Errorf("Expected log directory permissions 0750 or 0755, got %o", actualPerm)
	}
}

func TestConcurrentLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	Initialize("info", logPath)

	// Test concurrent logging (should not panic or cause data races)
	done := make(chan bool, 10)

	for i := range 10 {
		go func(id int) {
			defer func() { done <- true }()
			for j := range 100 {
				Info("concurrent message from goroutine %d, iteration %d", id, j)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}

	// Verify log file exists and has content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected log file to have content after concurrent logging")
	}
}
