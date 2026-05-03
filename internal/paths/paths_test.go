package paths

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRepoDir(t *testing.T) {
	home := homeDir()

	tests := []struct {
		name     string
		chroot   bool
		expected string
	}{
		{"normal", false, home + "/src/zerno"},
		{"chroot", true, "/mnt/root/src/zerno"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RepoDir(tt.chroot)
			if got != tt.expected {
				t.Errorf("RepoDir() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSrcDir(t *testing.T) {
	home := homeDir()

	tests := []struct {
		name     string
		chroot   bool
		expected string
	}{
		{"normal", false, home + "/src"},
		{"chroot", true, "/mnt/root/src"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SrcDir(tt.chroot)
			if got != tt.expected {
				t.Errorf("SrcDir() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfDir(t *testing.T) {
	home := homeDir()

	tests := []struct {
		name     string
		chroot   bool
		expected string
	}{
		{"normal", false, home + "/.zerno"},
		{"chroot", true, "/mnt/root/.zerno"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConfDir(tt.chroot)
			if got != tt.expected {
				t.Errorf("ConfDir() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHomeDir(t *testing.T) {
	home := homeDir()
	if home == "" {
		t.Error("homeDir() should not Return empty string")
	}
	if _, err := os.Stat(home); os.IsNotExist(err) {
		t.Errorf("homeDir() = %v, should exist", home)
	}
}

func TestHostBinPath(t *testing.T) {
	path := HostBinPath()
	if path == "" {
		t.Error("HostBinPath() should not return empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("HostBinPath() = %q, want absolute path", path)
	}
}

func TestCurrentUser(t *testing.T) {
	user := CurrentUser()
	if user == "" {
		t.Error("CurrentUser() should not return empty string")
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, user)
	if !matched {
		t.Errorf("CurrentUser() = %q, want valid username format", user)
	}
}

func TestIsoBuildsDir(t *testing.T) {
	path := IsoBuildsDir()
	if path == "" {
		t.Error("IsoBuildsDir() should not return empty string")
	}
	if !strings.HasSuffix(path, "/src/zerno-iso-builds") {
		t.Errorf("IsoBuildsDir() = %q, want suffix /src/zerno-iso-builds", path)
	}
}

func TestIsoMountDir(t *testing.T) {
	path := IsoMountDir()
	if path == "" {
		t.Error("IsoMountDir() should not return empty string")
	}
	if !strings.HasSuffix(path, "/src/zerno-iso-mnt") {
		t.Errorf("IsoMountDir() = %q, want suffix /src/zerno-iso-mnt", path)
	}
}

func TestRepoSrcDir(t *testing.T) {
	path := RepoSrcDir()
	if path == "" {
		t.Error("RepoSrcDir() should not return empty string")
	}
	if !strings.HasSuffix(path, "/src/zerno") {
		t.Errorf("RepoSrcDir() = %q, want suffix /src/zerno", path)
	}
}
