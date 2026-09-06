package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigSaveLoad(t *testing.T) {
	cfg := &Config{
		BlockDevice:   "sda",
		PartNum:       2,
		PartNumPrefix: "",
		Timezone:      "America/New_York",
		Hostname:      "testhost",
		Username:      "testuser",
		UserID:        1000,
		UserGID:       1000,
		NetDev:        "enp0s3",
		NetDevISO:     "eth0",
		WiFiEnabled:   true,
		WiFiSSID:      "MyNetwork",
		WiFiPassword:  "secret",
	}

	dir := t.TempDir()
	paramsFile := filepath.Join(dir, "parameters.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paramsFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	loadedData, err := os.ReadFile(paramsFile)
	if err != nil {
		t.Fatal(err)
	}

	var loaded Config
	if err := json.Unmarshal(loadedData, &loaded); err != nil {
		t.Fatal(err)
	}

	if loaded.Hostname != cfg.Hostname {
		t.Errorf("Hostname = %v, want %v", loaded.Hostname, cfg.Hostname)
	}
	if loaded.WiFiSSID != cfg.WiFiSSID {
		t.Errorf("WiFiSSID = %v, want %v", loaded.WiFiSSID, cfg.WiFiSSID)
	}
}

func TestValidateStrictCpuGovernor(t *testing.T) {
	base := &Config{
		Hostname:    "testhost",
		Username:    "testuser",
		BlockDevice: "sda",
		Timezone:    "UTC",
		NetDev:      "enp0s3",
	}
	base.CpuGovernor = "powersave"
	if err := base.ValidateStrict(); err != nil {
		t.Errorf("valid governor rejected: %v", err)
	}
	base.CpuGovernor = "performance"
	if err := base.ValidateStrict(); err != nil {
		t.Errorf("valid governor rejected: %v", err)
	}
	base.CpuGovernor = "userspace;rm"
	if err := base.ValidateStrict(); err == nil {
		t.Error("shell-hostile governor accepted")
	}
	base.CpuGovernor = ""
	if err := base.ValidateStrict(); err != nil {
		t.Errorf("empty governor should be allowed (old parameters.json): %v", err)
	}
}

func TestConfigLoad_FileNotFound(t *testing.T) {
	paramsFileOverride = filepath.Join(t.TempDir(), "parameters.json")
	t.Cleanup(func() { paramsFileOverride = "" })

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for nonexistent file")
	}
}

func TestConfigSaveLoad_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	paramsFileOverride = filepath.Join(dir, "parameters.json")
	t.Cleanup(func() { paramsFileOverride = "" })

	cfg := &Config{
		BlockDevice:   "nvme0n1",
		PartNum:       2,
		PartNumPrefix: "p",
		Timezone:      "Europe/Berlin",
		Hostname:      "testhost",
		Username:      "testuser",
		UserID:        1000,
		UserGID:       1000,
		NetDev:        "enp3s0",
		NetDevISO:     "eth0",
		WiFiEnabled:   true,
		WiFiSSID:      "Home Network",
		WiFiPassword:  "hunter2",
		SecureBoot:    true,
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	info, err := os.Stat(paramsFileOverride)
	if err != nil {
		t.Fatalf("Save() did not write params file: %v", err)
	}
	if info.IsDir() {
		t.Error("params path should be a regular file")
	}
	if got := info.Mode().Perm(); got != 0640 {
		t.Errorf("params file mode = %v, want 0640 (contains wifi passphrase)", got)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Hostname != cfg.Hostname {
		t.Errorf("Hostname = %q, want %q", loaded.Hostname, cfg.Hostname)
	}
	if loaded.WiFiSSID != cfg.WiFiSSID {
		t.Errorf("WiFiSSID = %q, want %q", loaded.WiFiSSID, cfg.WiFiSSID)
	}
	if loaded.WiFiPassword != cfg.WiFiPassword {
		t.Errorf("WiFiPassword = %q, want %q", loaded.WiFiPassword, cfg.WiFiPassword)
	}
	if loaded.NetDev != cfg.NetDev {
		t.Errorf("NetDev = %q, want %q", loaded.NetDev, cfg.NetDev)
	}
	if loaded.SecureBoot != cfg.SecureBoot {
		t.Errorf("SecureBoot = %v, want %v", loaded.SecureBoot, cfg.SecureBoot)
	}
}

func TestLoad_DefaultsSecureBootFalse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "parameters.json")
	paramsFileOverride = path
	t.Cleanup(func() { paramsFileOverride = "" })

	// pre-existing parameters.json without the SecureBoot field must load
	// as false (backward compatibility)
	data, err := json.Marshal(&Config{
		BlockDevice: "sda",
		PartNum:     2,
		Hostname:    "h",
		Username:    "u",
		NetDev:      "enp3s0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.SecureBoot {
		t.Error("SecureBoot should default to false for legacy parameters.json")
	}
}

func TestLoad_RejectsIncompleteConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "parameters.json")
	paramsFileOverride = path
	t.Cleanup(func() { paramsFileOverride = "" })

	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Error("Load() should reject config missing required fields")
	}
}

func TestConfigString(t *testing.T) {
	cfg := &Config{
		Hostname: "test",
	}

	s := cfg.String()
	if s == "" {
		t.Error("String() should not return empty string")
	}
}

func TestGetConfigDir(t *testing.T) {
	dir := getConfigDir()
	if dir == "" {
		t.Error("getConfigDir() should not return empty string")
	}

	home, _ := os.UserHomeDir()
	if home != "" && dir != filepath.Join(home, ".zerno") {
		t.Errorf("getConfigDir() = %v, want %v/.zerno", dir, home)
	}
}

func TestValidate(t *testing.T) {
	valid := Config{
		Hostname:    "dhost",
		BlockDevice: "sda",
		Username:    "user",
		PartNum:     2,
		NetDev:      "enp0s3",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("fully valid config rejected: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*Config)
	}{
		{"hostname required", func(c *Config) { c.Hostname = "" }},
		{"block device required", func(c *Config) { c.BlockDevice = "" }},
		{"username required", func(c *Config) { c.Username = "" }},
		{"partnum zero", func(c *Config) { c.PartNum = 0 }},
		{"partnum negative", func(c *Config) { c.PartNum = -1 }},
		{"netdev required", func(c *Config) { c.NetDev = "" }},
	}
	for _, tc := range cases {
		cfg := valid
		tc.mutate(&cfg)
		if err := cfg.Validate(); err == nil {
			t.Errorf("%s: expected error", tc.name)
		}
	}
}

func TestGetParametersFile_DefaultPath(t *testing.T) {
	paramsFileOverride = ""
	t.Cleanup(func() { paramsFileOverride = "" })
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SUDO_USER", "")
	got := getParametersFile()
	want := filepath.Join(home, ".zerno", "parameters.json")
	if got != want {
		t.Errorf("getParametersFile() = %q, want %q", got, want)
	}
}

func TestGetParametersFile_Override(t *testing.T) {
	path := filepath.Join(t.TempDir(), "params.json")
	paramsFileOverride = path
	t.Cleanup(func() { paramsFileOverride = "" })
	if got := getParametersFile(); got != path {
		t.Errorf("getParametersFile() = %q, want %q", got, path)
	}
}

func TestConfigPartialFields(t *testing.T) {
	cfg := &Config{
		WiFiEnabled:  false,
		WiFiSSID:     "",
		WiFiPassword: "",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if loaded.WiFiEnabled != false {
		t.Error("WiFiEnabled should be false")
	}
	if loaded.WiFiSSID != "" {
		t.Error("WiFiSSID should be empty")
	}
	if loaded.WiFiPassword != "" {
		t.Error("WiFiPassword should be empty")
	}
}

func TestConfigJSONRoundtrip(t *testing.T) {
	cfg := &Config{
		BlockDevice:   "nvme0n1",
		PartNum:       2,
		PartNumPrefix: "p",
		Timezone:      "Europe/Berlin",
		Hostname:      "testhost",
		Username:      "testuser",
		UserID:        1000,
		UserGID:       1000,
		NetDev:        "enp0s1",
		NetDevISO:     "wlan0",
		WiFiEnabled:   true,
		WiFiSSID:      "TestNetwork",
		WiFiPassword:  "secret123",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if loaded.BlockDevice != cfg.BlockDevice {
		t.Errorf("BlockDevice = %v, want %v", loaded.BlockDevice, cfg.BlockDevice)
	}
	if loaded.Hostname != cfg.Hostname {
		t.Errorf("Hostname = %v, want %v", loaded.Hostname, cfg.Hostname)
	}
	if loaded.WiFiSSID != cfg.WiFiSSID {
		t.Errorf("WiFiSSID = %v, want %v", loaded.WiFiSSID, cfg.WiFiSSID)
	}
	if loaded.WiFiPassword != cfg.WiFiPassword {
		t.Errorf("WiFiPassword = %v, want %v", loaded.WiFiPassword, cfg.WiFiPassword)
	}
}

func TestParseBoolOr(t *testing.T) {
	tests := []struct {
		input string
		def   bool
		want  bool
	}{
		{"y", false, true},
		{"Yes", true, true},
		{"TRUE", false, true},
		{"1", false, true},
		{"n", true, false},
		{"No", true, false},
		{"0", true, false},
		{"", true, true},       // empty -> default
		{"", false, false},     // empty -> default
		{"banana", true, true}, // garbage -> default (safe direction)
		{"banana", false, false},
	}
	for _, tt := range tests {
		if got := parseBoolOr(tt.input, tt.def); got != tt.want {
			t.Errorf("parseBoolOr(%q, %v) = %v, want %v", tt.input, tt.def, got, tt.want)
		}
	}
}

func TestPartNumPrefix(t *testing.T) {
	tests := []struct {
		dev  string
		want string
	}{
		{"sda", ""},
		{"vdb", ""},
		{"hda", ""},
		{"nvme0n1", "p"},
		{"nvme1n1", "p"},
		{"mmcblk0", "p"},
		{"mmcblk1", "p"},
	}
	for _, tt := range tests {
		if got := partNumPrefix(tt.dev); got != tt.want {
			t.Errorf("partNumPrefix(%q) = %q, want %q", tt.dev, got, tt.want)
		}
	}
}

func TestValidateStrict(t *testing.T) {
	valid := Config{
		Hostname:    "dhost",
		BlockDevice: "sda",
		Username:    "user",
		PartNum:     2,
		NetDev:      "enp3s0",
		Timezone:    "Asia/Singapore",
	}
	if err := valid.ValidateStrict(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*Config)
	}{
		{"hostname space", func(c *Config) { c.Hostname = "my host" }},
		{"hostname shell chars", func(c *Config) { c.Hostname = "host$(reboot)" }},
		{"username space", func(c *Config) { c.Username = "my user" }},
		{"username semicolon", func(c *Config) { c.Username = "user;x" }},
		{"block device path", func(c *Config) { c.BlockDevice = "sda;rm" }},
		{"block device slash", func(c *Config) { c.BlockDevice = "a/b" }},
		{"timezone space", func(c *Config) { c.Timezone = "Mars Olympus" }},
	}
	for _, tc := range cases {
		cfg := valid
		tc.mutate(&cfg)
		if err := cfg.ValidateStrict(); err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}

	// accepted variants that must NOT fail (lax rules: no false positives)
	good := []struct {
		name   string
		mutate func(*Config)
	}{
		{"single-char hostname", func(c *Config) { c.Hostname = "x" }},
		{"hyphenated username", func(c *Config) { c.Username = "john-doe_2" }},
		{"uppercase username", func(c *Config) { c.Username = "User" }},
		{"long hostname", func(c *Config) { c.Hostname = strings.Repeat("a", 100) }},
		{"dotted hostname", func(c *Config) { c.Hostname = "arch.local" }},
		{"underscore hostname", func(c *Config) { c.Hostname = "my_host" }},
		{"nvme device", func(c *Config) { c.BlockDevice = "nvme0n1" }},
		{"unknown timezone", func(c *Config) { c.Timezone = "Mars/Olympus" }},
		{"utc offset timezone", func(c *Config) { c.Timezone = "Etc/UTC+5" }},
		{"wifi enabled ok", func(c *Config) {
			c.WiFiEnabled = true
			c.WiFiSSID = "Home Network"
			c.WiFiPassword = "hunter2"
		}},
		{"open network no password", func(c *Config) {
			c.WiFiEnabled = true
			c.WiFiSSID = "FreeWiFi"
			c.WiFiPassword = ""
		}},
	}
	for _, tc := range good {
		cfg := valid
		tc.mutate(&cfg)
		if err := cfg.ValidateStrict(); err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
		}
	}

}
