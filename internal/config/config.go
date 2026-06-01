package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

	if err := checkUEFI(); err != nil {
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

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	fmt.Printf("\nparameters:\n%s\n", cfg)

	if !confirm("proceed with the installation?") {
		os.Exit(0)
	}

	return cfg, nil
}

func checkUEFI() error {
	if _, err := os.Stat("/sys/firmware/efi"); os.IsNotExist(err) {
		return fmt.Errorf("systemd-boot requires UEFI — /sys/firmware/efi not found")
	}
	return nil
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
	fmt.Scanln(&cfg.BlockDevice)
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
	var wifiStr string
	fmt.Scanln(&wifiStr)
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
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	*value = strings.TrimSpace(input)
	if *value == "" {
		*value = fallback
	}
}

func promptChoice(label string, options []string) string {
	fmt.Printf("%s %v: ", label, options)
	var choice string
	fmt.Scanln(&choice)
	if choice == "" && len(options) > 0 {
		return options[0]
	}
	return choice
}

func confirm(msg string) bool {
	for {
		fmt.Printf("%s [yn] ", msg)
		var input string
		fmt.Scanln(&input)
		switch input {
		case "y", "Y":
			return true
		case "n", "N":
			return false
		default:
			fmt.Printf("unknown input '%s', please enter y or n\n", input)
		}
	}
}
