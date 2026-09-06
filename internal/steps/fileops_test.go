package steps

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyRecursive_SingleFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "file.txt")
	dst := filepath.Join(t.TempDir(), "copy.txt")
	content := "hello"

	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := CopyRecursive(src, dst); err != nil {
		t.Fatalf("CopyRecursive() error = %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("content = %q, want %q", string(got), content)
	}
}

func TestCopyRecursive_SingleFilePermissions(t *testing.T) {
	src := filepath.Join(t.TempDir(), "script.sh")
	dst := filepath.Join(t.TempDir(), "script.sh")

	if err := os.WriteFile(src, []byte("#!/bin/sh"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := CopyRecursive(src, dst); err != nil {
		t.Fatalf("CopyRecursive() error = %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("permissions = %v, want 0755", info.Mode().Perm())
	}
}

func TestCopyRecursive_SingleDir(t *testing.T) {
	src := filepath.Join(t.TempDir(), "mydir")
	dst := filepath.Join(t.TempDir(), "copydir")

	// Create source directory with files and subdirs
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("b"), 0644)

	if err := CopyRecursive(src, dst); err != nil {
		t.Fatalf("CopyRecursive() error = %v", err)
	}

	// Verify structure
	tests := []struct {
		path    string
		isDir   bool
		content string
	}{
		{path: dst, isDir: true},
		{path: filepath.Join(dst, "a.txt"), content: "a"},
		{path: filepath.Join(dst, "sub"), isDir: true},
		{path: filepath.Join(dst, "sub", "b.txt"), content: "b"},
	}
	for _, tc := range tests {
		info, err := os.Stat(tc.path)
		if err != nil {
			t.Errorf("stat %s: %v", tc.path, err)
			continue
		}
		if tc.isDir && !info.IsDir() {
			t.Errorf("%s: expected directory", tc.path)
		}
		if tc.content != "" {
			data, _ := os.ReadFile(tc.path)
			if string(data) != tc.content {
				t.Errorf("%s: content = %q, want %q", tc.path, string(data), tc.content)
			}
		}
	}
}

func TestCopyRecursive_EmptyDir(t *testing.T) {
	src := filepath.Join(t.TempDir(), "empty")
	dst := filepath.Join(t.TempDir(), "copy")

	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := CopyRecursive(src, dst); err != nil {
		t.Fatalf("CopyRecursive() error = %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestCopyRecursive_MissingSrc(t *testing.T) {
	err := CopyRecursive("/nonexistent/path", "/tmp/dest")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestCopyRecursive_CreatesParentDirs(t *testing.T) {
	src := filepath.Join(t.TempDir(), "file.txt")
	dst := filepath.Join(t.TempDir(), "a", "b", "file.txt")

	if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := CopyRecursive(src, dst); err != nil {
		t.Fatalf("CopyRecursive() error = %v", err)
	}
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Error("CopyRecursive() did not create parent directories")
	}
}

func TestCopyRecursive_DirPermissions(t *testing.T) {
	src := filepath.Join(t.TempDir(), "restricted")
	dst := filepath.Join(t.TempDir(), "copy")

	if err := os.MkdirAll(src, 0700); err != nil {
		t.Fatal(err)
	}
	if err := CopyRecursive(src, dst); err != nil {
		t.Fatalf("CopyRecursive() error = %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("permissions = %v, want 0700", info.Mode().Perm())
	}
}

func TestFileExists(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "test.txt")
	_, err := os.Create(tmpfile)
	if err != nil {
		t.Fatal(err)
	}

	if !FileExists(tmpfile) {
		t.Errorf("FileExists(%q) = false, want true", tmpfile)
	}
	if FileExists("/nonexistent/path/to/file") {
		t.Errorf("FileExists(nonexistent) = true, want false")
	}
}

func TestWriteReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "hello world\nline 2"

	if err := WriteFile(path, content); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got != content {
		t.Errorf("ReadFile() = %q, want %q", got, content)
	}
}

func TestWriteFile_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "subdir", "nested")
	path := filepath.Join(dir, "test.txt")

	if err := WriteFile(path, "content"); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if !FileExists(path) {
		t.Errorf("WriteFile should create parent directories")
	}
}

func TestCopyFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source.txt")
	dst := filepath.Join(t.TempDir(), "dest.txt")
	content := "copy me"

	os.WriteFile(src, []byte(content), 0644)

	if err := CopyFile(src, dst, 0644); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	got, _ := ReadFile(dst)
	if got != content {
		t.Errorf("CopyFile() content = %q, want %q", got, content)
	}
}

func TestSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")

	os.WriteFile(target, []byte("target"), 0644)

	if err := Symlink(target, link); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	if !FileExists(link) {
		t.Errorf("Symlink() did not create link")
	}

	// Calling again should not fail (link already exists)
	if err := Symlink(target, link); err != nil {
		t.Errorf("Symlink() should not fail on existing link: %v", err)
	}
}

func TestLineInFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	// File doesn't exist - should create with line
	err := LineInFile(path, "first line")
	if err != nil {
		t.Fatalf("LineInFile() error = %v", err)
	}

	content, _ := ReadFile(path)
	if content != "first line\n" {
		t.Errorf("LineInFile() created wrong content: %q", content)
	}

	// Add another line
	err = LineInFile(path, "second line")
	if err != nil {
		t.Fatalf("LineInFile() error = %v", err)
	}

	content, _ = ReadFile(path)
	if content != "first line\nsecond line\n" {
		t.Errorf("LineInFile() content = %q", content)
	}

	// Adding duplicate should not add again
	err = LineInFile(path, "first line")
	if err != nil {
		t.Fatalf("LineInFile() error = %v", err)
	}

	// Count occurrences
	content, _ = ReadFile(path)
	count := strings.Count(content, "first line\n")
	if count != 1 {
		t.Errorf("LineInFile() should not add duplicate, got count = %d", count)
	}
}

func TestReplaceLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	content := `line1
# comment
line3
# another comment`

	os.WriteFile(path, []byte(content), 0644)

	// Replace comment lines
	err := ReplaceLine(path, `^#.*`, "# replaced")
	if err != nil {
		t.Fatalf("ReplaceLine() error = %v", err)
	}

	result, _ := ReadFile(path)
	expected := `line1
# replaced
line3
# replaced`

	if result != expected {
		t.Errorf("ReplaceLine() = %q, want %q", result, expected)
	}
}

func TestReplaceLine_Specific(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	content := `hello world
foo bar
baz qux`

	os.WriteFile(path, []byte(content), 0644)

	// Replace only "foo bar" line
	err := ReplaceLine(path, `^foo bar$`, "hello world")
	if err != nil {
		t.Fatalf("ReplaceLine() error = %v", err)
	}

	result, _ := ReadFile(path)
	expected := `hello world
hello world
baz qux`

	if result != expected {
		t.Errorf("ReplaceLine() = %q, want %q", result, expected)
	}
}

func TestChmod(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpfile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	err := os.Chmod(tmpfile, 0755)
	if err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}

	info, err := os.Stat(tmpfile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode() != 0755 {
		t.Errorf("Chmod() mode = %v, want 0755", info.Mode())
	}
}

func TestSymlink_UpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	target1 := filepath.Join(dir, "target1")
	target2 := filepath.Join(dir, "target2")
	link := filepath.Join(dir, "link")

	if err := os.WriteFile(target1, []byte("target1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target2, []byte("target2"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(target1, link); err != nil {
		t.Fatal(err)
	}

	err := Symlink(target2, link)
	if err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	readlink, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	if readlink != target2 {
		t.Errorf("Symlink() link points to %v, want %v", readlink, target2)
	}
}

func TestCopyFile_CreatesDirs(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	dst := filepath.Join(dir, "nested/sub/dir/dest.txt")

	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	err := CopyFile(src, dst, 0644)
	if err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Error("CopyFile() did not create nested directories")
	}
}

func TestCopyFile_EnforcesPermOnExistingDst(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	// pre-existing destination with a permissive mode (simulates re-runs)
	if err := os.WriteFile(dst, []byte("old"), 0666); err != nil {
		t.Fatal(err)
	}

	if err := CopyFile(src, dst, 0640); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0640 {
		t.Errorf("CopyFile() mode = %v, want 0640 (final mode must be enforced)", info.Mode().Perm())
	}
}

func TestReplaceLine_NoMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	err := ReplaceLine(path, `^nonexistent.*$`, "replacement")
	if err != nil {
		t.Fatalf("ReplaceLine() error = %v", err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "hello world" {
		t.Errorf("ReplaceLine() modified file without match, got: %q", data)
	}
}

func TestReplaceLine_MultipleMatches(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(path, []byte("foo\nfoo\nfoo"), 0644); err != nil {
		t.Fatal(err)
	}

	err := ReplaceLine(path, `^foo$`, "bar")
	if err != nil {
		t.Fatalf("ReplaceLine() error = %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Count(string(data), "bar") != 3 {
		t.Errorf("ReplaceLine() should replace all matches, got: %q", data)
	}
}

func TestMove_SameFilesystem(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "original.txt")
	dst := filepath.Join(dir, "moved.txt")
	content := "hello"

	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Move(src, dst); err != nil {
		t.Fatalf("Move() error = %v", err)
	}

	if FileExists(src) {
		t.Error("Move() source still exists")
	}
	got, err := ReadFile(dst)
	if err != nil {
		t.Fatalf("Move() dst not readable: %v", err)
	}
	if got != content {
		t.Errorf("Move() content = %q, want %q", got, content)
	}
}

func TestMove_MissingSrc(t *testing.T) {
	err := Move(filepath.Join(t.TempDir(), "nope"), filepath.Join(t.TempDir(), "dest"))
	if err == nil {
		t.Error("Move() missing source should return error")
	}
}

func TestReadLine(t *testing.T) {
	old := stdin
	t.Cleanup(func() { stdin = old })

	stdin = bufio.NewReader(strings.NewReader("  hello  \n"))
	if got := ReadLine(); got != "hello" {
		t.Errorf("ReadLine() = %q, want %q", got, "hello")
	}

	stdin = bufio.NewReader(strings.NewReader("no newline"))
	if got := ReadLine(); got != "no newline" {
		t.Errorf("ReadLine() without trailing newline = %q, want %q", got, "no newline")
	}
}

func TestAskConfirmation_YesNoLoop(t *testing.T) {
	old := stdin
	t.Cleanup(func() { stdin = old })

	// invalid input then y
	stdin = bufio.NewReader(strings.NewReader("x\ny\n"))
	if !AskConfirmation("?") {
		t.Error("AskConfirmation() should return true for y")
	}

	// n
	stdin = bufio.NewReader(strings.NewReader("n\n"))
	if AskConfirmation("?") {
		t.Error("AskConfirmation() should return false for n")
	}
}

func TestRunFailure(t *testing.T) {
	cause := errors.New("exit status 1")
	err := runFailure("pacman -S --needed nginx", "error: target not found: nginx\n", cause)
	if err == nil {
		t.Fatal("runFailure() should return non-nil error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "pacman -S --needed nginx") {
		t.Errorf("error should contain command label: %v", msg)
	}
	if !strings.Contains(msg, "target not found") {
		t.Errorf("error should contain last output line: %v", msg)
	}
}
