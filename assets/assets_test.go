package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"zerno/internal/config"
)

func TestRestore(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "bashrc")

	err := Restore("files/bashrc", dst)
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Error("Restore() did not create file")
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	embedded, err := assetsDir.ReadFile("files/bashrc")
	if err != nil {
		t.Fatalf("ReadFile() embedded error = %v", err)
	}

	if string(data) != string(embedded) {
		t.Errorf("Restore() content mismatch")
	}
}

func TestRestoreTemplate(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "hosts")

	cfg := &config.Config{Hostname: "testhost"}

	err := RestoreTemplate("base/hosts.tpl", dst, cfg)
	if err != nil {
		t.Fatalf("RestoreTemplate() error = %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	content := string(data)

	if content == "" {
		t.Error("RestoreTemplate() produced empty file")
	}

	if strings.Contains(content, "{{.Hostname}}") {
		t.Error("RestoreTemplate() did not replace {{.Hostname}}")
	}

	if !strings.Contains(content, "testhost") {
		t.Errorf("RestoreTemplate() did not substitute hostname, got: %s", content)
	}
}

func TestRestoreTemplate_NilDataError(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "hosts")

	err := RestoreTemplate("base/hosts.tpl", dst, nil)
	if err == nil {
		t.Error("RestoreTemplate(nil) should return error")
	}
}

func TestRestore_NonexistentAsset(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "output")

	err := Restore("nonexistent/file", dst)
	if err == nil {
		t.Error("Restore() should return error for nonexistent asset")
	}
}

func TestRestoreTemplate_NonexistentAsset(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "output")

	err := RestoreTemplate("nonexistent/file", dst, &config.Config{})
	if err == nil {
		t.Error("RestoreTemplate() should return error for nonexistent asset")
	}
}

func TestRestore_CreatesDirs(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "nested/sub/dir/bashrc")

	err := Restore("files/bashrc", dst)
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Error("Restore() did not create nested directories")
	}
}
