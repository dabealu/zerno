package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"zerno/assets"
	"zerno/internal/config"
	"zerno/internal/paths"
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

func printError(name string, err error) {
	fmt.Printf(red+"▒ %s: %v"+reset+"\n", name, err)
	os.Exit(1)
}

func command(name, cmdStr string) Task {
	return Task{
		Name: name,
		RunFunc: func(cfg *config.Config) error {
			_, err := steps.RunShell(cmdStr)
			return err
		},
	}
}

func info(msg string) Task {
	return Task{
		Name: "info",
		RunFunc: func(cfg *config.Config) error {
			log.Println(msg)
			return nil
		},
	}
}

func requireUser(user string) Task {
	return Task{
		Name: "require_run_as",
		RunFunc: func(cfg *config.Config) error {
			fmt.Println("required user:", user)
			if user == "root" {
				if os.Getuid() != 0 {
					return fmt.Errorf("required user: root, current user: %s", paths.CurrentUser())
				}
				return nil
			}
			if paths.CurrentUser() != user {
				return fmt.Errorf("required user: %s, current user: %s", user, paths.CurrentUser())
			}
			return nil
		},
	}
}

func copyFile(assetPath, destPath string) Task {
	return Task{
		Name: "copy_file",
		RunFunc: func(cfg *config.Config) error {
			log.Printf("copying %s -> %s", assetPath, destPath)
			return assets.Restore(assetPath, destPath)
		},
	}
}

func copyTemplate(assetPath, destPath string, data any) Task {
	return Task{
		Name: "template_file",
		RunFunc: func(cfg *config.Config) error {
			log.Printf("rendering %s -> %s", assetPath, destPath)
			return assets.RestoreTemplate(assetPath, destPath, data)
		},
	}
}

func runTaskList(tasks []Task, cfg *config.Config) {
	total := len(tasks)
	for i, t := range tasks {
		printStep(i+1, total, t.Name)
		if err := t.RunFunc(cfg); err != nil {
			printError(t.Name, err)
		}
	}
}
