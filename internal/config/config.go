package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"zerno/internal/steps"
)

type Config struct {
	BlockDevice   string
	PartNum       int
	PartNumPrefix string
	Timezone      string
	Hostname      string
	Username      string
	UserID        int
	UserGID       int
	NetDev        string
	NetDevISO     string
	WiFiEnabled   bool
	WiFiSSID      string
	WiFiPassword  string
}

func (c *Config) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

func (c *Config) Validate() error {
	if c.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if c.BlockDevice == "" {
		return fmt.Errorf("block device is required")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.PartNum <= 0 {
		return fmt.Errorf("invalid partition number: %d", c.PartNum)
	}
	if c.NetDev == "" {
		return fmt.Errorf("network device is required")
	}
	return nil
}

// ValidateStrict applies basic sanity checks to freshly prompted input:
// each field must consist only of characters that are valid for it. This
// catches typos and shell-hostile garbage while staying lax enough to never
// block a reasonable install. Load() intentionally stays lenient so
// pre-existing parameters.json files never break re-runs of install-full.
func (c *Config) ValidateStrict() error {
	if !regexp.MustCompile(`^[a-zA-Z0-9._-]+$`).MatchString(c.Hostname) {
		return fmt.Errorf("invalid hostname %q: allowed: letters, digits, '.', '-', '_'", c.Hostname)
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9._-]+$`).MatchString(c.Username) {
		return fmt.Errorf("invalid username %q: allowed: letters, digits, '.', '-', '_'", c.Username)
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(c.BlockDevice) {
		return fmt.Errorf("invalid block device %q: allowed: letters and digits", c.BlockDevice)
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]+(?:[/_+-][a-zA-Z0-9]+)*$`).MatchString(c.Timezone) {
		return fmt.Errorf("invalid timezone %q: allowed: letters, digits, '/', '_', '+', '-'", c.Timezone)
	}
	return nil
}

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".zerno")
}

func getParametersFile() string {
	return filepath.Join(getConfigDir(), "parameters.json")
}

func (c *Config) Save() error {
	dir := getConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(getParametersFile(), data, 0644)
}

func Load() (*Config, error) {
	data, err := os.ReadFile(getParametersFile())
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}

func Prompt() (*Config, error) {
	cfg := &Config{}

	fmt.Println("enter parameters:")

	if err := CheckUEFI(); err != nil {
		return nil, err
	}
	if err := selectBlockDevice(cfg); err != nil {
		return nil, err
	}
	promptBasicInfo(cfg)
	if err := selectNetworkDevice(cfg); err != nil {
		return nil, err
	}
	promptWiFi(cfg)

	if err := cfg.ValidateStrict(); err != nil {
		return nil, err
	}

	fmt.Printf("\nparameters:\n%s\n", cfg)

	if !steps.AskConfirmation("proceed with the installation?") {
		os.Exit(0)
	}

	return cfg, nil
}

func selectBlockDevice(cfg *Config) error {
	devices, err := listBlockDevices()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return fmt.Errorf("no block devices found")
	}

	fmt.Println()
	lsblk, err := exec.Command("lsblk", "-o", "NAME,SIZE,TYPE,FSTYPE,MOUNTPOINT").Output()
	if err != nil {
		return fmt.Errorf("failed to list block devices: %w", err)
	}
	fmt.Print(string(lsblk))
	fmt.Println()

	fmt.Printf("block device %v: ", devices)
	cfg.BlockDevice = steps.ReadLine()
	if cfg.BlockDevice == "" {
		cfg.BlockDevice = devices[0]
	}

	if len(cfg.BlockDevice) >= 4 && cfg.BlockDevice[:4] == "nvme" {
		cfg.PartNumPrefix = "p"
	} else {
		cfg.PartNumPrefix = ""
	}
	cfg.PartNum = 2
	return nil
}

func promptBasicInfo(cfg *Config) {
	prompt("timezone", &cfg.Timezone, "Asia/Singapore")
	prompt("hostname", &cfg.Hostname, "dhost")
	prompt("username", &cfg.Username, "user")
	cfg.UserID = 1000
	cfg.UserGID = 1000
}

func selectNetworkDevice(cfg *Config) error {
	cfg.NetDevISO = promptNetDevice()
	if cfg.NetDevISO == "" {
		return fmt.Errorf("no network interface found")
	}

	cfg.NetDev = getNetDevName(cfg.NetDevISO)
	fmt.Printf("%s will be named %s after archiso\n", cfg.NetDevISO, cfg.NetDev)
	return nil
}

func promptWiFi(cfg *Config) {
	defaultWiFi := strings.HasPrefix(cfg.NetDevISO, "wlan") || strings.HasPrefix(cfg.NetDevISO, "wlp")
	cfg.WiFiEnabled = defaultWiFi

	fmt.Print("configure wifi [", defaultWiFi, "]: ")
	wifiStr := steps.ReadLine()
	if wifiStr != "" {
		cfg.WiFiEnabled = wifiStr == "true" || wifiStr == "1"
	}

	if cfg.WiFiEnabled {
		prompt("wifi ssid", &cfg.WiFiSSID, "")
		prompt("wifi password", &cfg.WiFiPassword, "")
	}
}

func LoadOrPrompt() (*Config, error) {
	cfg, err := Load()
	if err == nil {
		fmt.Println("got parameters from config file")
		return cfg, nil
	}

	cfg, err = Prompt()
	if err != nil {
		return nil, err
	}

	if err := cfg.Save(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func listBlockDevices() ([]string, error) {
	out, err := exec.Command("lsblk", "--output=NAME", "--noheadings", "--nodeps").Output()
	if err != nil {
		return nil, fmt.Errorf("list block devices: %w", err)
	}

	var devices []string
	for _, line := range strings.Fields(string(out)) {
		if line != "" {
			devices = append(devices, line)
		}
	}
	return devices, nil
}

func promptNetDevice() string {
	entries, err := os.ReadDir("/sys/class/net/")
	if err != nil {
		return ""
	}

	netInterfaceRegex := regexp.MustCompile(`^(wlan|wlp|eth|enp).*`)

	var devices []string
	for _, entry := range entries {
		if netInterfaceRegex.MatchString(entry.Name()) {
			devices = append(devices, entry.Name())
		}
	}

	if len(devices) == 0 {
		fmt.Println("no network interfaces found")
		return ""
	}

	return promptChoice("network interface", devices)
}

func getNetDevName(isoDev string) string {
	out, err := exec.Command("udevadm", "test-builtin", "net_id", filepath.Join("/sys/class/net", isoDev)).Output()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ID_NET_NAME_PATH=") {
			return strings.TrimPrefix(line, "ID_NET_NAME_PATH=")
		}
	}
	return ""
}

func prompt(label string, value *string, fallback string) {
	fmt.Printf("%s [%s]: ", label, fallback)
	*value = steps.ReadLine()
	if *value == "" {
		*value = fallback
	}
}

func promptChoice(label string, options []string) string {
	fmt.Printf("%s %v: ", label, options)
	choice := steps.ReadLine()
	if choice == "" && len(options) > 0 {
		return options[0]
	}
	return choice
}

// CheckUEFI reports whether the system booted in UEFI mode (required by systemd-boot).
func CheckUEFI() error {
	if _, err := os.Stat("/sys/firmware/efi"); os.IsNotExist(err) {
		return fmt.Errorf("systemd-boot requires UEFI — /sys/firmware/efi not found")
	}
	return nil
}
