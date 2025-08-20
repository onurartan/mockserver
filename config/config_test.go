// [WHY_ONLY_TEST_HAVE_CONFIG] => Only the config area was tested in the project because this part is important.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func createTempMockFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp mock file: %v", err)
	}
	return path
}

func TestValidateMock_Success(t *testing.T) {
	tmpDir := t.TempDir()
	mockFile := createTempMockFile(t, tmpDir, "todos.json", `{"todos": []}`)

	mockCfg := &MockConfig{
		File:    mockFile,
		Status:  200,
		DelayMs: 10,
	}

	err := validateMock(mockCfg, "/todos", "")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateMock_FileNotFound(t *testing.T) {
	mockCfg := &MockConfig{
		File:    "not_found.json",
		Status:  200,
		DelayMs: 0,
	}

	err := validateMock(mockCfg, "/todos", "")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestValidateMock_InvalidExtension(t *testing.T) {
	tmpDir := t.TempDir()
	mockFile := createTempMockFile(t, tmpDir, "todos.txt", `dummy`)

	mockCfg := &MockConfig{
		File: mockFile,
	}

	err := validateMock(mockCfg, "/todos", "")
	if err == nil {
		t.Error("expected error for invalid extension, got nil")
	}
}

func TestValidateAndApplyDefaults(t *testing.T) {
	cfg := &Config{}
	err := validateAndApplyDefaults(cfg, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 5000 {
		t.Errorf("expected default port=5000, got %d", cfg.Server.Port)
	}

	if cfg.Server.DefaultHeaders["Content-Type"] != "application/json" {
		t.Errorf("expected default header 'application/json', got %v", cfg.Server.DefaultHeaders)
	}
}
