package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makePCIDevice creates a fake sysfs pci device entry with the given class
// and vendor files.
func makePCIDevice(t *testing.T, root, slot, class, vendor string) {
	t.Helper()
	dir := filepath.Join(root, "bus", "pci", "devices", slot)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"class": class, "vendor": vendor} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetectGPUVendor(t *testing.T) {
	const (
		vgaIntel   = "0x030000"
		vgaAmd     = "0x030000"
		threeDAmd  = "0x030200"
		threeDNV   = "0x030200"
		netControl = "0x0c0310" // non-display class, must be ignored
	)

	tests := []struct {
		name    string
		devices func(*testing.T, string)
		want    string
		wantErr string
	}{
		{
			name: "intel igpu only",
			devices: func(t *testing.T, root string) {
				makePCIDevice(t, root, "0000:00:02.0", vgaIntel, pciVendorIntel)
			},
			want: "intel",
		},
		{
			name: "amd dgpu wins over intel igpu",
			devices: func(t *testing.T, root string) {
				makePCIDevice(t, root, "0000:00:02.0", vgaIntel, pciVendorIntel)
				makePCIDevice(t, root, "0000:01:00.0", threeDAmd, pciVendorAMD)
			},
			want: "amd",
		},
		{
			name: "amd apu only",
			devices: func(t *testing.T, root string) {
				makePCIDevice(t, root, "0000:05:00.0", vgaAmd, pciVendorAMD)
			},
			want: "amd",
		},
		{
			name: "optimus nvidia as 3d class is rejected",
			devices: func(t *testing.T, root string) {
				makePCIDevice(t, root, "0000:00:02.0", vgaIntel, pciVendorIntel)
				makePCIDevice(t, root, "0000:01:00.0", threeDNV, pciVendorNVIDIA)
			},
			wantErr: "nvidia gpus are not supported",
		},
		{
			name: "no display controllers",
			devices: func(t *testing.T, root string) {
				if err := os.MkdirAll(filepath.Join(root, "bus", "pci", "devices"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: "no supported gpu found",
		},
		{
			name: "non-display classes ignored",
			devices: func(t *testing.T, root string) {
				makePCIDevice(t, root, "0000:00:14.3", netControl, pciVendorIntel)
			},
			wantErr: "no supported gpu found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.devices(t, root)

			got, err := detectGPUVendor(root)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("detectGPUVendor() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("detectGPUVendor() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("detectGPUVendor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectGPUVendor_MissingSysFS(t *testing.T) {
	_, err := detectGPUVendor(filepath.Join(t.TempDir(), "nonexistent"))
	if err == nil {
		t.Error("detectGPUVendor() should fail when sysfs tree is missing")
	}
}
