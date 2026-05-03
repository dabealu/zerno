package steps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	got, _ := ReadFile(dst)
	if got != content {
		t.Errorf("CopyFile() content = %q, want %q", got, content)
	}
}

func TestCreateDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b", "c")

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("CreateDir() error = %v", err)
	}

	if !FileExists(dir) {
		t.Errorf("CreateDir() did not create directory")
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

func TestReplaceBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	content := `start
[old section]
block line 1
block line 2
[end section]
end`

	os.WriteFile(path, []byte(content), 0644)

	// Replace block between [old section] and [end section]
	err := ReplaceBlock(path, `\[old section\]`, `\[end section\]`, "new block content")
	if err != nil {
		t.Fatalf("ReplaceBlock() error = %v", err)
	}

	result, _ := ReadFile(path)
	expected := `start
[old section]
new block content
[end section]
end`

	if result != expected {
		t.Errorf("ReplaceBlock() = %q, want %q", result, expected)
	}
}

func TestReplaceBlock_NotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	content := `start
middle
end`

	os.WriteFile(path, []byte(content), 0644)

	// Block not found - should not modify
	err := ReplaceBlock(path, `\[nonexistent\]`, `\[end\]`, "replacement")
	if err != nil {
		t.Fatalf("ReplaceBlock() error = %v", err)
	}

	result, _ := ReadFile(path)
	if result != content {
		t.Errorf("ReplaceBlock() should not modify when block not found")
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

	err := CopyFile(src, dst)
	if err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Error("CopyFile() did not create nested directories")
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
