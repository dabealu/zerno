package install

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
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

func TestWifiDevice(t *testing.T) {
	tests := []struct {
		name   string
		isoDev string
		want   string
	}{
		{"wlan0", "wlan0", "wlan0"},
		{"wlp", "wlp3s0", "wlp3s0"},
		{"wlx mac", "wlx00aa11bb22cc", "wlx00aa11bb22cc"},
		{"ethernet falls back", "enp0s3", "wlan0"},
		{"empty falls back", "", "wlan0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wifiDevice(&config.Config{NetDevISO: tt.isoDev})
			if got != tt.want {
				t.Errorf("wifiDevice(%q) = %q, want %q", tt.isoDev, got, tt.want)
			}
		})
	}
}

func TestWifiConnectCmd(t *testing.T) {
	tests := []struct {
		name string
		dev  string
		ssid string
		pass string
		want []string
	}{
		{
			name: "with passphrase",
			dev:  "wlan0",
			ssid: "Home Network",
			pass: "secret",
			want: []string{"iwctl", "--passphrase", "secret", "station", "wlan0", "connect", "Home Network"},
		},
		{
			name: "open network no passphrase flag",
			dev:  "wlan0",
			ssid: "Cafe Wifi",
			pass: "",
			want: []string{"iwctl", "station", "wlan0", "connect", "Cafe Wifi"},
		},
		{
			name: "ssid with spaces is one argv element",
			dev:  "wlp2s0",
			ssid: "My Network Name",
			pass: "pw",
			want: []string{"iwctl", "--passphrase", "pw", "station", "wlp2s0", "connect", "My Network Name"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wifiConnectCmd(tt.dev, tt.ssid, tt.pass)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wifiConnectCmd(%q,%q,%q) = %v, want %v", tt.dev, tt.ssid, tt.pass, got, tt.want)
			}
		})
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
