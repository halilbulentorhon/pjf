package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/halilbulentorhon/pjf/internal/config"
	"github.com/halilbulentorhon/pjf/internal/ide"
	"github.com/halilbulentorhon/pjf/internal/scanner"
	"github.com/halilbulentorhon/pjf/internal/service"
	"github.com/halilbulentorhon/pjf/internal/tui"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "uninstall":
			uninstall()
			return
		case "update":
			selfUpdate()
			return
		case "--version", "-v":
			fmt.Println("pjf " + version)
			return
		}
	}

	configPath := config.DefaultPath()
	cachePath := config.DefaultCachePath()

	cfg, isFirstRun, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	sc := &scanner.FileScanner{}
	cache := &scanner.JSONCacheStore{Path: cachePath}
	svc := service.New(cfg, sc, cache)
	svc.SetDetectedIDEs(ide.DetectAll())

	updateCh := make(chan string, 1)
	go func() {
		updateCh <- checkForUpdate()
	}()

	if err := tui.Run(svc, configPath, isFirstRun); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	select {
	case latest := <-updateCh:
		if latest != "" {
			fmt.Printf("\nUpdate available: v%s → v%s\nRun: pjf update\n", version, latest)
		}
	default:
	}
}

func uninstall() {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "pjf")
	cacheDir := filepath.Join(home, ".cache", "pjf")

	fmt.Println("pjf will be uninstalled. The following will be deleted:")
	fmt.Printf("  %s\n", configDir)
	fmt.Printf("  %s\n", cacheDir)

	exe, err := os.Executable()
	if err == nil {
		exe, _ = filepath.EvalSymlinks(exe)
		fmt.Printf("  %s\n", exe)
	}

	fmt.Print("\nContinue? [y/N] ")
	exec.Command("stty", "-f", "/dev/tty", "cbreak", "min", "1").Run()
	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	exec.Command("stty", "-f", "/dev/tty", "-cbreak").Run()
	fmt.Printf("%c\n", buf[0])

	if buf[0] != 'y' && buf[0] != 'Y' {
		fmt.Println("Cancelled.")
		return
	}

	removed := 0
	if err := os.RemoveAll(configDir); err == nil {
		fmt.Printf("  Removed: %s\n", configDir)
		removed++
	}
	if err := os.RemoveAll(cacheDir); err == nil {
		fmt.Printf("  Removed: %s\n", cacheDir)
		removed++
	}

	if exe != "" {
		if err := os.Remove(exe); err == nil {
			fmt.Printf("  Removed: %s\n", exe)
			removed++
		} else {
			fmt.Printf("  Could not remove binary: %v\n", err)
			fmt.Printf("  Remove manually: rm %s\n", exe)
		}
	}

	fmt.Println("\npjf has been uninstalled.")
}

const repo = "halilbulentorhon/pjf"

type ghRelease struct {
	TagName string `json:"tag_name"`
}

func fetchLatestVersion() string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/" + repo + "/releases/latest")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ""
	}
	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return ""
	}
	return strings.TrimPrefix(rel.TagName, "v")
}

func checkForUpdate() string {
	if version == "dev" {
		return ""
	}
	latest := fetchLatestVersion()
	if latest == "" || latest == version {
		return ""
	}
	return latest
}

func selfUpdate() {
	fmt.Println("Checking for updates...")
	latest := fetchLatestVersion()
	if latest == "" {
		fmt.Println("Could not check for updates.")
		return
	}
	if version != "dev" && latest == version {
		fmt.Printf("Already up to date (v%s).\n", version)
		return
	}

	fmt.Printf("Updating pjf to v%s...\n", latest)

	osName := runtime.GOOS
	arch := runtime.GOARCH
	url := fmt.Sprintf("https://github.com/%s/releases/download/v%s/pjf_%s_%s.tar.gz", repo, latest, osName, arch)

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not find current binary: %v\n", err)
		return
	}
	exe, _ = filepath.EvalSymlinks(exe)

	tmp, err := os.MkdirTemp("", "pjf-update-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tmp)

	dl := exec.Command("sh", "-c", fmt.Sprintf("curl -sSL '%s' | tar xz -C '%s'", url, tmp))
	if err := dl.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		return
	}

	newBin := filepath.Join(tmp, "pjf")
	if _, err := os.Stat(newBin); err != nil {
		fmt.Fprintf(os.Stderr, "Downloaded binary not found\n")
		return
	}

	cp := exec.Command("sh", "-c", fmt.Sprintf("sudo cp '%s' '%s'", newBin, exe))
	if err := cp.Run(); err != nil {
		mv := exec.Command("cp", newBin, exe)
		if err := mv.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Could not replace binary: %v\n", err)
			fmt.Printf("Try manually: sudo cp %s %s\n", newBin, exe)
			return
		}
	}

	fmt.Printf("Updated to v%s.\n", latest)
}
