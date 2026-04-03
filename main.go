package main

import (
	"fmt"
	"os"

	"github.com/halilbulentorhon/pjf/internal/config"
	"github.com/halilbulentorhon/pjf/internal/scanner"
	"github.com/halilbulentorhon/pjf/internal/service"
	"github.com/halilbulentorhon/pjf/internal/tui"
)

func main() {
	configPath := config.DefaultPath()
	cachePath := config.DefaultCachePath()

	cfg, isFirstRun, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config yüklenemedi: %v\n", err)
		os.Exit(1)
	}

	sc := &scanner.FileScanner{}
	cache := &scanner.JSONCacheStore{Path: cachePath}
	svc := service.New(cfg, sc, cache)

	if err := tui.Run(svc, configPath, isFirstRun); err != nil {
		fmt.Fprintf(os.Stderr, "Hata: %v\n", err)
		os.Exit(1)
	}
}
