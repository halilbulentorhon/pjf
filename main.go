package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/term"

	"github.com/halilbulentorhon/pjf/internal/config"
	"github.com/halilbulentorhon/pjf/internal/ide"
	"github.com/halilbulentorhon/pjf/internal/scanner"
	"github.com/halilbulentorhon/pjf/internal/service"
	"github.com/halilbulentorhon/pjf/internal/tui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "uninstall" {
		uninstall()
		return
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

	if err := tui.Run(svc, configPath, isFirstRun); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
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
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("\nCannot read input.")
		return
	}
	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	term.Restore(int(os.Stdin.Fd()), oldState)
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
