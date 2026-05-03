package steps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration_LineInFile_MultipleAdds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	lines := []string{"line1", "line2", "line3"}
	for _, l := range lines {
		if err := LineInFile(path, l); err != nil {
			t.Fatalf("LineInFile(%q) error = %v", l, err)
		}
	}

	content, _ := ReadFile(path)
	for _, l := range lines {
		if !strings.Contains(content, l) {
			t.Errorf("LineInFile should add %q to file", l)
		}
	}

	// Adding same lines again should not duplicate
	for _, l := range lines {
		if err := LineInFile(path, l); err != nil {
			t.Fatalf("LineInFile duplicate %q error = %v", l, err)
		}
	}

	content, _ = ReadFile(path)
	count := strings.Count(content, "line1\n")
	if count != 1 {
		t.Errorf("LineInFile should not duplicate lines, got count=%d, want 1", count)
	}
}

func TestIntegration_ReplaceLine_RegexPatterns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	content := `# Config file
DEBUG=true
LOG_LEVEL=info
# DEBUG=false
PROD=true
`

	os.WriteFile(path, []byte(content), 0644)

	// Replace all DEBUG lines (both uncommented and commented)
	if err := ReplaceLine(path, `^DEBUG=.*`, "DEBUG=false"); err != nil {
		t.Fatalf("ReplaceLine error = %v", err)
	}

	result, _ := ReadFile(path)

	// Count how many DEBUG= lines are now "false"
	countFalse := strings.Count(result, "DEBUG=false")
	countTrue := strings.Count(result, "DEBUG=true")

	if countFalse < 1 {
		t.Error("ReplaceLine should add DEBUG=false")
	}
	if countTrue > 0 {
		t.Error("ReplaceLine should remove DEBUG=true")
	}
}

func TestIntegration_ReplaceBlock_ConfigSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	content := `[Package]
Name=test
Version=1.0

[Settings]
enabled=false
timeout=30

[Other]
key=value`

	os.WriteFile(path, []byte(content), 0644)

	// Replace entire [Settings] block
	if err := ReplaceBlock(path, `\[Settings\]`, `\[Other\]`, "[Settings]\nenabled=true\ntimeout=60"); err != nil {
		t.Fatalf("ReplaceBlock error = %v", err)
	}

	result, _ := ReadFile(path)

	if !strings.Contains(result, "enabled=true") {
		t.Error("ReplaceBlock should add enabled=true")
	}
	if !strings.Contains(result, "timeout=60") {
		t.Error("ReplaceBlock should add timeout=60")
	}
	if strings.Contains(result, "enabled=false") {
		t.Error("ReplaceBlock should remove old enabled=false")
	}
	if strings.Contains(result, "timeout=30") {
		t.Error("ReplaceBlock should remove old timeout=30")
	}
}

func TestIntegration_CopyFile_PreservesContent(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	dst := filepath.Join(dir, "subdir", "dest.txt")

	complexContent := `# Sway config
set $mod Mod4
bindsym $mod+Return exec alacritty
bindsym $mod+d exec dmenu_run

exec_always waybar
`

	os.WriteFile(src, []byte(complexContent), 0644)

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile error = %v", err)
	}

	result, _ := ReadFile(dst)
	if result != complexContent {
		t.Errorf("CopyFile content mismatch:\ngot:  %q\nwant: %q", result, complexContent)
	}
}

func TestIntegration_Symlink_CreatesLink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "binary")
	link := filepath.Join(dir, "link")

	os.WriteFile(target, []byte("data"), 0644)

	if err := Symlink(target, link); err != nil {
		t.Fatalf("Symlink error = %v", err)
	}

	// Verify link exists and points to target
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("Link should exist: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Link should be a symlink")
	}

	// Read through symlink
	data, _ := os.ReadFile(link)
	if string(data) != "data" {
		t.Errorf("Reading through symlink should work, got %q", string(data))
	}
}

func TestIntegration_WriteFile_NestedDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "d", "file.txt")
	content := "nested file content"

	if err := WriteFile(path, content); err != nil {
		t.Fatalf("WriteFile with nested dirs error = %v", err)
	}

	if !FileExists(path) {
		t.Error("WriteFile should create nested directories")
	}

	result, _ := ReadFile(path)
	if result != content {
		t.Errorf("Written content mismatch: got %q, want %q", result, content)
	}
}

func TestIntegration_FileOperations_Chain(t *testing.T) {
	dir := t.TempDir()

	// Simulate a config file lifecycle
	configPath := filepath.Join(dir, "config.txt")

	// 1. Create initial file
	WriteFile(configPath, "initial=value\n")

	// 2. Add lines
	LineInFile(configPath, "debug=true")
	LineInFile(configPath, "log=info")

	// 3. Replace a line
	ReplaceLine(configPath, `^initial=.*`, "initial=updated")

	// 4. Verify final state
	content, _ := ReadFile(configPath)

	if !strings.Contains(content, "initial=updated") {
		t.Error("final content should have updated initial value")
	}
	if !strings.Contains(content, "debug=true") {
		t.Error("final content should have debug=true")
	}
	if !strings.Contains(content, "log=info") {
		t.Error("final content should have log=info")
	}
}

func TestIntegration_NonExistentFile_Read(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.txt")

	_, err := ReadFile(path)
	if err == nil {
		t.Error("ReadFile should return error for nonexistent file")
	}
}

func TestIntegration_NonExistentFile_ReplaceLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")

	// ReplaceLine on non-existent file should return error
	err := ReplaceLine(path, `.*`, "replacement")
	if err == nil {
		t.Error("ReplaceLine should return error for nonexistent file")
	}
	if FileExists(path) {
		t.Error("ReplaceLine should not create file if it doesn't exist")
	}
}

func TestIntegration_CopyFile_NonExistentSource(t *testing.T) {
	src := filepath.Join(t.TempDir(), "nonexistent")
	dst := filepath.Join(t.TempDir(), "dest.txt")

	err := CopyFile(src, dst)
	if err == nil {
		t.Error("CopyFile should return error when source doesn't exist")
	}
}

func TestRunCmd_Echo(t *testing.T) {
	out, err := RunCmd("echo", "hello", "world")
	if err != nil {
		t.Fatalf("RunCmd() error = %v", err)
	}
	if !strings.Contains(out, "hello world") {
		t.Errorf("RunCmd() output = %q, want contains 'hello world'", out)
	}
}

func TestRunShell_Echo(t *testing.T) {
	out, err := RunShell("echo hello from shell")
	if err != nil {
		t.Fatalf("RunShell() error = %v", err)
	}
	if !strings.Contains(out, "hello from shell") {
		t.Errorf("RunShell() output = %q", out)
	}
}

func TestRunCmd_NotFound(t *testing.T) {
	_, err := RunCmd("nonexistent_command_xyz_12345")
	if err == nil {
		t.Error("RunCmd() should return error for nonexistent command")
	}
}

func TestRunShell_InvalidCommand(t *testing.T) {
	_, err := RunShell("exit 1")
	if err == nil {
		t.Error("RunShell() should return error for failing command")
	}
}

func TestRunCmd_WithArgs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := RunCmd("cat", path)
	if err != nil {
		t.Fatalf("RunCmd() error = %v", err)
	}
	if !strings.Contains(out, "test content") {
		t.Errorf("RunCmd() output = %q, want 'test content'", out)
	}
}

func TestRunShell_WithArgs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(path, []byte("line1\nline2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := RunShell("wc -l < " + path)
	if err != nil {
		t.Fatalf("RunShell() error = %v", err)
	}
	if !strings.Contains(out, "2") {
		t.Errorf("RunShell() wc -l output = %q, want contains '2'", out)
	}
}
