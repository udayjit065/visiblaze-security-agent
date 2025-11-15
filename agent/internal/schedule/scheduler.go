package schedule

import (
	"time"

	"github.com/visiblaze/sec-agent/agent/internal/cis"
	"github.com/visiblaze/sec-agent/agent/internal/collect"
	"github.com/visiblaze/sec-agent/agent/internal/config"
	"github.com/visiblaze/sec-agent/agent/internal/ingest"
	"github.com/visiblaze/sec-agent/agent/internal/logging"
)

type Scheduler struct {
	cfg    *config.Config
	logger *logging.Logger
	ticker *time.Ticker
	done   chan struct{}
}

func New(cfg *config.Config, logger *logging.Logger) *Scheduler {
	return &Scheduler{
		cfg:    cfg,
		logger: logger,
		done:   make(chan struct{}),
	}
}

func (s *Scheduler) RunOnce() error {
	return s.collect()
}

func (s *Scheduler) Start() {
	s.ticker = time.NewTicker(time.Duration(s.cfg.CollectionIntervalMinutes) * time.Minute)
	defer s.ticker.Stop()

	if err := s.collect(); err != nil {
		s.logger.Errorf("Initial collection failed: %v", err)
	}

	for {
		select {
		case <-s.ticker.C:
			if err := s.collect(); err != nil {
				s.logger.Errorf("Scheduled collection failed: %v", err)
			}
		case <-s.done:
			return
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.done)
}

func (s *Scheduler) collect() error {
	hostInfo, err := collect.GetHostInfo("0.1.0")
	if err != nil {
		s.logger.Errorf("Failed to collect host info: %v", err)
		return err
	}

	packages, _ := collect.CollectPackages(hostInfo.OSID)
	cisResults := cis.RunAllChecks()

	payload := map[string]interface{}{
		"host":        hostInfo,
		"packages":    packages,
		"cis_results": cisResults,
	}

	client := ingest.NewClient(s.cfg, s.logger)
	if err := client.SendPayload(payload); err != nil {
		s.logger.Errorf("Failed to send payload: %v", err)
		return err
	}

	s.logger.Infof("Collection complete")
	return nil
}
