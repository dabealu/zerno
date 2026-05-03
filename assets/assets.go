package assets

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed base conf files qemu sysctl.d utilsfs
var assetsDir embed.FS

func Restore(path, dst string) error {
	log.Printf("restoring %s -> %s", path, dst)
	data, err := assetsDir.ReadFile(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func RestoreTemplate(path, dst string, data any) error {
	log.Printf("rendering template %s -> %s", path, dst)
	content, err := assetsDir.ReadFile(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	if data == nil {
		return fmt.Errorf("data parameter cannot be nil")
	}
	tmpl, err := template.New(path).Parse(string(content))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}
	return os.WriteFile(dst, buf.Bytes(), 0644)
}
