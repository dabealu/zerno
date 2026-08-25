package install

import (
	"strings"
	"testing"

	"zerno/internal/config"
)

func TestRootPartitionPath(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{"sata", &config.Config{BlockDevice: "sda", PartNumPrefix: "", PartNum: 2}, "/dev/sda2"},
		{"mmc", &config.Config{BlockDevice: "mmcblk0", PartNumPrefix: "p", PartNum: 2}, "/dev/mmcblk0p2"},
		{"nvme", &config.Config{BlockDevice: "nvme0n1", PartNumPrefix: "p", PartNum: 2}, "/dev/nvme0n1p2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rootPartitionPath(tt.cfg); got != tt.want {
				t.Errorf("rootPartitionPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBaseKernelCmdline(t *testing.T) {
	got := baseKernelCmdline("abc-123")
	want := "loglevel=6 root=UUID=abc-123\n"
	if got != want {
		t.Errorf("baseKernelCmdline() = %q, want %q", got, want)
	}
}

func TestHibernationKernelCmdline(t *testing.T) {
	got := hibernationKernelCmdline("root-u", "swap-u", "38400")
	want := "loglevel=6 root=UUID=root-u resume=UUID=swap-u resume_offset=38400\n"
	if got != want {
		t.Errorf("hibernationKernelCmdline() = %q, want %q", got, want)
	}
}

// Boot-critical strings are embedded constants; these assertions catch
// accidental drift when editing bootloader().
func TestBootloaderAssets(t *testing.T) {
	if !strings.Contains(loaderConf, "default arch-linux*") {
		t.Error("loader.conf must default to the UKI glob pattern")
	}
	for _, hook := range []string{"systemd", "microcode", "kms", "keyboard", "sd-vconsole", "filesystems"} {
		if !strings.Contains(initramfsHooks, hook) {
			t.Errorf("initramfs hooks missing %q", hook)
		}
	}
	if !strings.Contains(linuxPreset, `default_uki="/efi/EFI/Linux/arch-linux.efi"`) {
		t.Error("linux.preset must build the UKI at /efi/EFI/Linux/arch-linux.efi")
	}
	for _, want := range []string{"When = PreTransaction", "arch-linux-fallback.efi"} {
		if !strings.Contains(preserveOldUKIHook, want) {
			t.Errorf("preserve-old-uki hook missing %q", want)
		}
	}
}
