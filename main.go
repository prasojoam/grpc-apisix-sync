package main

import (
	"flag"
	"fmt"
	"github.com/prasojoam/grpc-apisix-sync/internal/apisix"
	"github.com/prasojoam/grpc-apisix-sync/internal/config"
	"github.com/prasojoam/grpc-apisix-sync/internal/sync"
	"log"
)

const Version = "v1.1.0"

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config file")
	dataPath := flag.String("data", "data.yaml", "Path to data file")

	showVersion := flag.Bool("version", false, "Show version information")

	flag.Parse()

	if *showVersion {
		fmt.Printf("grpc-apisix-sync %s\n", Version)
		return
	}

	// 1. Load Files
	cfg, data, err := config.LoadFiles(*configPath, *dataPath)
	if err != nil {
		log.Fatalf("Failed to load files: %v", err)
	}

	// 2. Initialize APISIX Client
	client := apisix.NewClient(cfg.Apisix.URL, cfg.Apisix.Key)

	// 3. Initialize and run Syncer
	syncer := sync.NewSyncer(cfg, data, client)

	if err := syncer.Sync(); err != nil {
		log.Fatalf("Sync failed: %v", err)
	}

	fmt.Println("\n🚀 Synchronization complete!")
}
