package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigSaveLoad(t *testing.T) {
	cfg := &Config{
		EFI:           true,
		BlockDevice:   "sda",
		PartNum:       2,
		PartNumPrefix: "",
		Timezone:      "America/New_York",
		Hostname:      "testhost",
		Username:      "testuser",
		UserID:        "1000",
		UserGID:       "1000",
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

	if loaded.EFI != cfg.EFI {
		t.Errorf("EFI = %v, want %v", loaded.EFI, cfg.EFI)
	}
	if loaded.Hostname != cfg.Hostname {
		t.Errorf("Hostname = %v, want %v", loaded.Hostname, cfg.Hostname)
	}
	if loaded.WiFiSSID != cfg.WiFiSSID {
		t.Errorf("WiFiSSID = %v, want %v", loaded.WiFiSSID, cfg.WiFiSSID)
	}
}

func TestConfigLoad_FileNotFound(t *testing.T) {
	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for nonexistent file")
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
		EFI:           true,
		BlockDevice:   "nvme0n1",
		PartNum:       2,
		PartNumPrefix: "p",
		Timezone:      "Europe/Berlin",
		Hostname:      "testhost",
		Username:      "testuser",
		UserID:        "1000",
		UserGID:       "1000",
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
