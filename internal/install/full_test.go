package install

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"zerno/assets"
	"zerno/internal/config"
)

func TestInstallSwayFiles(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{
		Username: "tester",
		UserID:   os.Getuid(),
		UserGID:  os.Getegid(),
	}

	if err := installSwayFiles(cfg, tmp); err != nil {
		t.Fatalf("installSwayFiles() error = %v", err)
	}

	for src, name := range swayConfAssets {
		dst := filepath.Join(tmp, ".config/sway", name)
		data, err := os.ReadFile(dst)
		if err != nil {
			t.Errorf("restored %s: %v", src, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("restored %s: empty file", src)
		}
		want, err := assets.ReadFile(src)
		if err != nil {
			t.Fatalf("read embedded %s: %v", src, err)
		}
		if !bytes.Equal(data, want) {
			t.Errorf("restored %s: content mismatch with embedded asset", src)
		}
	}

	for _, f := range swayExecutables {
		info, err := os.Stat(filepath.Join(tmp, f))
		if err != nil {
			t.Errorf("executable %s: %v", f, err)
			continue
		}
		if info.Mode()&0o111 == 0 {
			t.Errorf("%s: not executable, mode %v", f, info.Mode())
		}
	}

	if _, err := os.Stat(filepath.Join(tmp, ".config/ghostty/config")); err != nil {
		t.Errorf("ghostty config: %v", err)
	}
}

func TestSwayExecutablesAreRestoredScripts(t *testing.T) {
	for _, f := range swayExecutables {
		base := filepath.Base(f)
		found := false
		for _, name := range swayConfAssets {
			if name == base {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("swayExecutables entry %q has no matching swayConfAssets entry", f)
		}
		if !strings.HasSuffix(base, ".sh") && !strings.HasSuffix(base, ".py") {
			t.Errorf("swayExecutables entry %q is not a script", f)
		}
	}
}
