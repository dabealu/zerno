package task

import (
	"fmt"
	"log"
	"os"
	"time"

	"zerno/assets"
	"zerno/internal/config"
	"zerno/internal/steps"
)

type Task struct {
	Name    string
	RunFunc func(cfg *config.Config) error
}

const (
	reset = "\033[0m"
	green = "\033[92m"
	red   = "\033[91m"
)

func printStep(current, total int, name string) {
	fmt.Printf(green+"▒ %d/%d | %s | %s"+reset+"\n", current, total, time.Now().Format("15:04:05"), name)
}

func Command(name, cmdStr string) Task {
	return Task{
		Name: name,
		RunFunc: func(cfg *config.Config) error {
			_, err := steps.RunShell(cmdStr)
			return err
		},
	}
}

func Pacman(name string, pkgs []string) Task {
	return Task{
		Name: name,
		RunFunc: func(cfg *config.Config) error {
			return steps.PacmanPackages(pkgs)
		},
	}
}

func Info(msg string) Task {
	return Task{
		Name: "info",
		RunFunc: func(cfg *config.Config) error {
			log.Println(msg)
			return nil
		},
	}
}

func RequireUser(user string) Task {
	return Task{
		Name: "require_run_as",
		RunFunc: func(cfg *config.Config) error {
			fmt.Println("required user:", user)
			if user == "root" {
				if os.Getuid() != 0 {
					return fmt.Errorf("required user: root, current user: %s", currentUser())
				}
				return nil
			}
			current := currentUser()
			if current != user {
				return fmt.Errorf("required user: %s, current user: %s", user, current)
			}
			return nil
		},
	}
}

func currentUser() string {
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		return sudoUser
	}
	return os.Getenv("USER")
}

func CopyFile(assetPath, destPath string) Task {
	return Task{
		Name: "copy_file",
		RunFunc: func(cfg *config.Config) error {
			log.Printf("copying %s -> %s", assetPath, destPath)
			return assets.Restore(assetPath, destPath)
		},
	}
}

func CopyTemplate(assetPath, destPath string, data any) Task {
	return Task{
		Name: "template_file",
		RunFunc: func(cfg *config.Config) error {
			log.Printf("rendering %s -> %s", assetPath, destPath)
			return assets.RestoreTemplate(assetPath, destPath, data)
		},
	}
}

func RunTaskList(tasks []Task, cfg *config.Config) error {
	total := len(tasks)
	for i, t := range tasks {
		printStep(i+1, total, t.Name)
		if err := t.RunFunc(cfg); err != nil {
			return fmt.Errorf("%s: %w", t.Name, err)
		}
	}
	return nil
}
