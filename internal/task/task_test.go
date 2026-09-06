package task

import (
	"errors"
	"os"
	"strings"
	"testing"

	"zerno/internal/config"
)

func TestRunTaskList_ExecutesInOrder(t *testing.T) {
	var calls []string
	tasks := []Task{
		{Name: "first", RunFunc: func(cfg *config.Config) error {
			calls = append(calls, "first")
			return nil
		}},
		{Name: "second", RunFunc: func(cfg *config.Config) error {
			calls = append(calls, "second")
			return nil
		}},
		{Name: "third", RunFunc: func(cfg *config.Config) error {
			calls = append(calls, "third")
			return nil
		}},
	}
	if err := RunTaskList(tasks, &config.Config{}); err != nil {
		t.Fatalf("RunTaskList() error = %v", err)
	}
	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d: %v", len(calls), calls)
	}
	want := []string{"first", "second", "third"}
	if !reflectSlice(calls, want) {
		t.Errorf("got %v, want %v", calls, want)
	}
}

func TestRunTaskList_StopsOnError(t *testing.T) {
	var calls []string
	tasks := []Task{
		{Name: "ok", RunFunc: func(cfg *config.Config) error {
			calls = append(calls, "ok")
			return nil
		}},
		{Name: "bad", RunFunc: func(cfg *config.Config) error {
			return errors.New("boom")
		}},
		{Name: "never", RunFunc: func(cfg *config.Config) error {
			calls = append(calls, "never")
			return nil
		}},
	}
	err := RunTaskList(tasks, &config.Config{})
	if err == nil {
		t.Fatal("expected error from bad task")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Errorf("error should contain task name 'bad': %v", err)
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("error should contain cause 'boom': %v", err)
	}
	if len(calls) != 1 || calls[0] != "ok" {
		t.Errorf("expected only [ok] before failure, got %v", calls)
	}
}

func TestRequireUser_MatchesCurrentUser(t *testing.T) {
	t.Setenv("SUDO_USER", "")
	t.Setenv("USER", "myuser")
	tk := RequireUser("myuser")
	if err := tk.RunFunc(&config.Config{}); err != nil {
		t.Errorf("RequireUser(current user) should succeed: %v", err)
	}
}

func TestRequireUser_Mismatch(t *testing.T) {
	t.Setenv("SUDO_USER", "")
	t.Setenv("USER", "myuser")
	tk := RequireUser("other")
	if err := tk.RunFunc(&config.Config{}); err == nil {
		t.Error("RequireUser(mismatch) should fail")
	}
}

func TestRequireUser_SudoUserTakesPrecedence(t *testing.T) {
	t.Setenv("SUDO_USER", "sudoer")
	t.Setenv("USER", "regular")
	tk := RequireUser("sudoer")
	if err := tk.RunFunc(&config.Config{}); err != nil {
		t.Errorf("RequireUser via SUDO_USER should succeed: %v", err)
	}
}

func TestCurrentUser_PrefersSudoUser(t *testing.T) {
	t.Setenv("SUDO_USER", "sudoer")
	t.Setenv("USER", "regular")
	if got := currentUser(); got != "sudoer" {
		t.Errorf("currentUser() = %q, want %q", got, "sudoer")
	}
}

func TestCurrentUser_FallsBackToUser(t *testing.T) {
	t.Setenv("SUDO_USER", "")
	t.Setenv("USER", "regular")
	if got := currentUser(); got != "regular" {
		t.Errorf("currentUser() = %q, want %q", got, "regular")
	}
}

func TestRequireUser_RootRequiresUidZero(t *testing.T) {
	if os.Getuid() == 0 {
		tk := RequireUser("root")
		if err := tk.RunFunc(&config.Config{}); err != nil {
			t.Errorf("RequireUser(root) should succeed as root: %v", err)
		}
	} else {
		tk := RequireUser("root")
		if err := tk.RunFunc(&config.Config{}); err == nil {
			t.Error("RequireUser(root) should fail as non-root")
		}
	}
}

func reflectSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
