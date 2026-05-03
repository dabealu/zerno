package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration_SaveLoad_JSON(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		EFI:           true,
		BlockDevice:   "nvme0n1",
		PartNum:       2,
		PartNumPrefix: "p",
		Timezone:      "Europe/Berlin",
		Hostname:      "myhost",
		Username:      "admin",
		UserID:        "1000",
		UserGID:       "1000",
		NetDev:        "enp0s1",
		NetDevISO:     "eth0",
		WiFiEnabled:   true,
		WiFiSSID:      "MyNetwork",
		WiFiPassword:  "secret123",
	}

	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	loadedData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(loadedData, &loaded); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	if loaded.EFI != cfg.EFI {
		t.Errorf("EFI mismatch: got %v, want %v", loaded.EFI, cfg.EFI)
	}
	if loaded.BlockDevice != cfg.BlockDevice {
		t.Errorf("BlockDevice mismatch: got %v, want %v", loaded.BlockDevice, cfg.BlockDevice)
	}
	if loaded.Hostname != cfg.Hostname {
		t.Errorf("Hostname mismatch: got %v, want %v", loaded.Hostname, cfg.Hostname)
	}
	if loaded.WiFiPassword != cfg.WiFiPassword {
		t.Errorf("WiFiPassword mismatch: got %v, want %v", loaded.WiFiPassword, cfg.WiFiPassword)
	}
}

func TestIntegration_SaveLoad_WiFiDisabled(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		WiFiEnabled:  false,
		WiFiSSID:     "",
		WiFiPassword: "",
	}

	path := filepath.Join(dir, "wifi_disabled.json")
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}
	os.WriteFile(path, data, 0644)

	loadedData, _ := os.ReadFile(path)
	var loaded Config
	json.Unmarshal(loadedData, &loaded)

	if loaded.WiFiEnabled != false {
		t.Errorf("WiFiEnabled should be false")
	}
	if loaded.WiFiSSID != "" {
		t.Errorf("WiFiSSID should be empty")
	}
}

func TestIntegration_ConfigString_Output(t *testing.T) {
	cfg := &Config{
		Hostname: "testhost",
		Username: "testuser",
	}

	s := cfg.String()

	if !strings.Contains(s, "testhost") {
		t.Error("String() output should contain hostname")
	}
	if !strings.Contains(s, "testuser") {
		t.Error("String() output should contain username")
	}
}

func TestIntegration_MultipleConfigs(t *testing.T) {
	dir := t.TempDir()

	configs := []*Config{
		{Hostname: "host1", Username: "user1"},
		{Hostname: "host2", Username: "user2"},
		{Hostname: "host3", Username: "user3"},
	}

	for i, cfg := range configs {
		path := filepath.Join(dir, "config_"+string(rune('a'+i))+".json")
		data, _ := json.Marshal(cfg)
		os.WriteFile(path, data, 0644)

		loadedData, _ := os.ReadFile(path)
		var loaded Config
		json.Unmarshal(loadedData, &loaded)

		if loaded.Hostname != cfg.Hostname {
			t.Errorf("Config %d: hostname mismatch", i)
		}
	}
}
