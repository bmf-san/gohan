package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/bmf-san/gohan/internal/config"
	"github.com/bmf-san/gohan/internal/server"
)

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	port := fs.Int("port", 1313, "port to listen on")
	host := fs.String("host", "127.0.0.1", "host/address to bind")
	configPath := fs.String("config", "config.yaml", "path to config file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Run an initial full build before starting the server
	fmt.Println("serve: running initial build...")
	if err := runBuild([]string{"--full", "--config", *configPath}); err != nil {
		// Non-fatal: warn but continue so the user can fix content while the server is running
		fmt.Printf("serve: initial build warning: %v\n", err)
	}

	// Determine project root and output directory from config.
	cfgAbs, err := filepath.Abs(*configPath)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}
	rootDir := filepath.Dir(cfgAbs)

	// Load config to get the actual output directory; fall back to "public".
	outDir := filepath.Join(rootDir, "public")
	if cfg, cfgErr := config.New(rootDir).Load(); cfgErr == nil {
		outDir = filepath.Join(rootDir, cfg.Build.OutputDir)
	}

	// rebuildFn triggers a differential build on file change
	rebuildFn := func() error {
		return runBuild([]string{"--config", *configPath})
	}

	srv := server.NewDevServer(*host, *port, outDir, rebuildFn)
	srv.RootDir = rootDir // resolve watch dirs relative to project root (M-6)
	fmt.Printf("serve: listening on http://%s:%d\n", *host, *port)
	return srv.Start()
}
