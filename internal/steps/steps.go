package steps

import (
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

func CopyFile(from, to string) error {
	log.Printf("copying %s -> %s", from, to)
	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	dir := filepath.Dir(to)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(to, data, 0644)
}

// CopyRecursive recursively copies a file or directory tree from src to dst.
// Regular files are copied with their source permissions. Symlinks are followed.
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

// ChownRecursive recursively changes the owner and group of path and all its contents.
func ChownRecursive(path string, uid, gid int) error {
	log.Printf("chown -R %d:%d %s", uid, gid, path)
	return filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
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

func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func Symlink(origin, link string) error {
	os.Remove(link)
	return os.Symlink(origin, link)
}

func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

func RunCmd(name string, args ...string) (string, error) {
	log.Printf("cmd: %s %s", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w", string(out), err)
	}
	return string(out), nil
}

func RunShell(script string) (string, error) {
	log.Printf("shell: %s", script)
	cmd := exec.Command("bash", "-o", "pipefail", "-ec", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w", string(out), err)
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

func AskConfirmation(msg string) bool {
	for {
		fmt.Printf("%s [yn] ", msg)
		var input string
		fmt.Scanln(&input)
		switch strings.ToLower(input) {
		case "y", "yes":
			return true
		case "n", "no":
			fmt.Println("exiting...")
			os.Exit(0)
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
