package paths

import (
	"os"
	"os/user"
	"path/filepath"
)

const (
	RepoURL  = "https://github.com/dabealu/zerno.git"
	RepoPath = "src/zerno"
	BinPath  = "/usr/local/bin/zerno"
)

func RepoDir(chroot bool) string {
	return filepath.Join(baseDir(chroot), RepoPath)
}

func SrcDir(chroot bool) string {
	return filepath.Join(baseDir(chroot), "src")
}

func ConfDir(chroot bool) string {
	return filepath.Join(baseDir(chroot), ".zerno")
}

func HostBinPath() string {
	exe, err := os.Executable()
	if err != nil {
		return BinPath
	}
	return exe
}

func CurrentUser() string {
	if sudUser := os.Getenv("SUDO_USER"); sudUser != "" {
		return sudUser
	}
	return os.Getenv("USER")
}

func baseDir(chroot bool) string {
	if chroot {
		return "/mnt/root"
	}
	return homeDir()
}

func homeDir() string {
	if sudUser := os.Getenv("SUDO_USER"); sudUser != "" {
		if usr, err := user.Lookup(sudUser); err == nil {
			return usr.HomeDir
		}
	}
	if home, _ := os.UserHomeDir(); home != "" {
		return home
	}
	return os.Getenv("HOME")
}

func IsoBuildsDir() string {
	return filepath.Join(SrcDir(false), "zerno-iso-builds")
}

func IsoMountDir() string {
	return filepath.Join(SrcDir(false), "zerno-iso-mnt")
}

func RepoSrcDir() string {
	return filepath.Join(os.Getenv("HOME"), RepoPath)
}
