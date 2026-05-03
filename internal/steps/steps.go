package steps

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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
	log.Printf("cmd: %s %v", name, args)
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w", string(out), err)
	}
	return string(out), nil
}

func RunShell(script string) (string, error) {
	log.Printf("shell: %s", script)
	cmd := exec.Command("bash", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w", string(out), err)
	}
	return string(out), nil
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

func ReplaceBlock(path, start, end, replacement string) error {
	log.Printf("replacing block in %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	startRe := regexp.MustCompile(start)
	endRe := regexp.MustCompile(end)

	lines := strings.Split(string(data), "\n")
	var newLines []string
	foundBlock := false
	inBlock := false

	for _, line := range lines {
		if !inBlock && startRe.MatchString(line) {
			newLines = append(newLines, line)
			inBlock = true
			continue
		}
		if inBlock {
			if endRe.MatchString(line) {
				newLines = append(newLines, replacement)
				newLines = append(newLines, line)
				inBlock = false
				foundBlock = true
				continue
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	if !foundBlock {
		return nil
	}
	return WriteFile(path, strings.Join(newLines, "\n"))
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
