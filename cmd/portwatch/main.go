// main is the entry point for the portwatch CLI daemon.
// It loads configuration, initializes components, and runs the monitoring loop.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/monitor"
	"github.com/user/portwatch/internal/state"
)

func main() {
	configPath := flag.String("config", "", "Path to JSON config file (optional)")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println("portwatch v0t	os.Exit(0)
	}

	// Load configuration — fall back to defaults if no file is specified.
	var cfg config.Config
	var err error
	if *configPath != "" {
		cfg, err = config.Load(*configPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
	} else {
		cfg = config.Default()
	}

	log.Printf("portwatch starting — scan interval: %s, ports: %v", cfg.Interval, cfg.Ports)

	// Set up the alerter to write to stdout.
	alerter := alert.New(os.Stdout)

	// Set up persistent state so we survive restarts.
	st := state.New(cfg.StateFile)

	// Build the monitor with loaded config and dependencies.
	mon := monitor.New(cfg, st, alerter)

	// Perform an initial scan immediately on startup.
	if err := mon.Run(); err != nil {
		log.Printf("initial scan error: %v", err)
	}

	// Ticker drives subsequent scans at the configured interval.
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	// Handle SIGINT / SIGTERM for a clean shutdown.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			if err := mon.Run(); err != nil {
				log.Printf("scan error: %v", err)
			}
		case sig := <-sigs:
			log.Printf("received signal %s — shutting down", sig)
			return
		}
	}
}
