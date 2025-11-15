package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/visiblaze/sec-agent/agent/internal/config"
	"github.com/visiblaze/sec-agent/agent/internal/logging"
	"github.com/visiblaze/sec-agent/agent/internal/schedule"
)

var Version = "0.1.0"

func main() {
	configPath := flag.String("config", "/etc/visiblaze-agent/config.yaml", "Path to config file")
	runOnce := flag.Bool("once", false, "Run collection once and exit")
	flag.Parse()

	// Initialize logging (allow override via VISIBLAZE_LOG_DIR for local dev)
	logDir := os.Getenv("VISIBLAZE_LOG_DIR")
	if logDir == "" {
		logDir = "/var/log/visiblaze-agent"
	}
	logger, err := logging.New(logDir)
	if err != nil {
		os.Stderr.WriteString("Failed to initialize logging: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer logger.Close()

	logger.Infof("Visiblaze Agent v%s starting", Version)

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Errorf("Failed to load config: %v", err)
		os.Exit(1)
	}

	// Initialize scheduler
	sched := schedule.New(cfg, logger)

	if *runOnce {
		logger.Infof("Running collection once")
		if err := sched.RunOnce(); err != nil {
			logger.Errorf("Collection failed: %v", err)
			os.Exit(1)
		}
		logger.Infof("Collection complete")
		return
	}

	// Start scheduler
	logger.Infof("Starting scheduler (interval: %d minutes)", cfg.CollectionIntervalMinutes)
	go sched.Start()

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Infof("Shutdown signal received")
	sched.Stop()
	logger.Infof("Agent stopped")
}
