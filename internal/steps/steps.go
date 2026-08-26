package steps

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func WriteFile(path, content string) error {
	log.Printf("writing %s", path)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// CopyFile copies from to to, enforcing perm as the final mode (applied even
// when the destination already exists - re-runs must not keep stale modes).
func CopyFile(from, to string, perm os.FileMode) error {
	log.Printf("copying %s -> %s (%o)", from, to, perm)
	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(to, data, perm); err != nil {
		return err
	}
	return os.Chmod(to, perm)
}

// CopyRecursive recursively copies a file or directory tree from src to dst.
// Regular files are copied with their source permissions. Symlinks are
// recreated as-is (not followed); absolute targets may dangle until the
// pointed-to files exist at the destination root.
func CopyRecursive(src, dst string) error {
	log.Printf("copy %s -> %s", src, dst)
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)
		info, err := d.Info()
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}
		if d.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(target, destPath)
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		dstFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			srcFile.Close()
			return err
		}
		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()
		return err
	})
}

// ChownRecursive recursively changes the owner and group of path and all its
// contents. Symlinks are skipped: chown(2) dereferences them, so walking a
// tree containing links (e.g. ~/de -> /usr/local/bin/de) would otherwise
// change ownership of targets outside the tree. Link ownership itself is
// irrelevant on Linux.
func ChownRecursive(path string, uid, gid int) error {
	log.Printf("chown -R %d:%d %s", uid, gid, path)
	return filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		return os.Chown(p, uid, gid)
	})
}

// Move moves src to dst, handling cross-device moves.
// First tries os.Rename (fast atomic move within the same filesystem).
// If that fails with a cross-device error (EXDEV), falls back to
// CopyRecursive + RemoveAll, which works across filesystem boundaries.
func Move(src, dst string) error {
	log.Printf("move %s -> %s", src, dst)
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if !errors.Is(err, syscall.EXDEV) {
		return err
	}
	if err := CopyRecursive(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func Symlink(origin, link string) error {
	os.Remove(link)
	return os.Symlink(origin, link)
}

func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

// errorLogFile collects full output of failed commands so the returned error
// can stay short while details remain available for debugging.
const errorLogFile = "/tmp/zerno-install.log"

// runFailure appends the failing command and its full output to the log file
// and returns a compact error carrying the last output lines.
func runFailure(label, out string, cause error) error {
	if f, ferr := os.OpenFile(errorLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); ferr == nil {
		fmt.Fprintf(f, "\n=== %s ===\n$ %s\n%s\n", time.Now().Format("2006-01-02 15:04:05"), label, out)
		f.Close()
	}

	// single-line tail of the last few raw output lines - usually identifies
	// the failure without opening the log; %q keeps it one clean line
	lines := strings.Split(strings.TrimSpace(out), "\n")
	const maxTailLines = 5
	if len(lines) > maxTailLines {
		lines = lines[len(lines)-maxTailLines:]
	}
	if len(lines) == 1 && lines[0] == "" {
		return fmt.Errorf("%s failed: %w", label, cause)
	}
	return fmt.Errorf("%s failed: %w; last output: %q (full output: %s)",
		label, cause, strings.Join(lines, " | "), errorLogFile)
}

func RunCmd(name string, args ...string) (string, error) {
	label := strings.Join(append([]string{name}, args...), " ")
	log.Printf("cmd: %s", label)
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return string(out), runFailure(label, string(out), err)
	}
	return string(out), nil
}

// RunCmdIn runs a command with working directory dir (same semantics as RunCmd).
func RunCmdIn(dir, name string, args ...string) (string, error) {
	label := fmt.Sprintf("(in %s) %s", dir, strings.Join(append([]string{name}, args...), " "))
	log.Printf("cmd: %s", label)
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), runFailure(label, string(out), err)
	}
	return string(out), nil
}

func RunShell(script string) (string, error) {
	log.Printf("shell: %s", script)
	out, err := exec.Command("bash", "-o", "pipefail", "-ec", script).CombinedOutput()
	if err != nil {
		return string(out), runFailure(script, string(out), err)
	}
	return string(out), nil
}

func PacmanPackages(pkgs []string) error {
	_, err := RunShell("pacman -Sy --noconfirm " + strings.Join(pkgs, " "))
	return err
}

func LineInFile(path, line string) error {
	if !FileExists(path) {
		log.Printf("creating %s with line: %s", path, line)
		return WriteFile(path, line+"\n")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, l := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(l) == line {
			log.Printf("skipped %q - already in %s", line, path)
			return nil
		}
	}
	log.Printf("appending %q to %s", line, path)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, line)
	return err
}

func ReplaceLine(path, pattern, replacement string) error {
	log.Printf("replacing in %s: %s -> %s", path, pattern, replacement)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(pattern)
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if re.MatchString(line) {
			lines[i] = re.ReplaceAllString(line, replacement)
		}
	}
	return WriteFile(path, strings.Join(lines, "\n"))
}

// stdin is the single shared buffered reader for all interactive input.
// Multiple readers on os.Stdin would steal buffered bytes from each other.
var stdin = bufio.NewReader(os.Stdin)

// ReadLine reads one line from stdin without the trailing newline.
// Input is trimmed of surrounding whitespace, matching old Scanln behavior.
func ReadLine() string {
	line, _ := stdin.ReadString('\n')
	return strings.TrimSpace(line)
}

func AskConfirmation(msg string) bool {
	for {
		fmt.Printf("%s [yn] ", msg)
		input := ReadLine()
		switch strings.ToLower(input) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Printf("unknown input '%s', please enter y or n\n", input)
		}
	}
}

func WaitForDefaultRoute(timeout int) error {
	log.Printf("waiting for default route to come up...")
	for range timeout {
		time.Sleep(1 * time.Second)
		out, _ := RunCmd("ip", "route", "show", "default")
		if strings.TrimSpace(out) != "" {
			return nil
		}
	}
	return fmt.Errorf("timeout: no default route after %d seconds", timeout)
}
