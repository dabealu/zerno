package testutils

import (
	"os"
	"testing"
)

func TempFile(t *testing.T, dir, pattern string) (*os.File, func()) {
	t.Helper()
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	return f, func() { os.Remove(f.Name()) }
}

func TempDir(t *testing.T, pattern string) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	return dir, func() { os.RemoveAll(dir) }
}

func WriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
