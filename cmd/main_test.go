package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintHelp(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	printHelp()
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	s := buf.String()

	for _, want := range []string{"install-base", "install-full", "build-iso", "boot-dev", "version", "readme"} {
		if !strings.Contains(s, want) {
			t.Errorf("printHelp() missing %q", want)
		}
	}
	for _, removed := range []string{"qemu", "steam", "update-bin", "repo-pull"} {
		if strings.Contains(s, removed) {
			t.Errorf("printHelp() still lists removed command %q", removed)
		}
	}
}
