package ingest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/visiblaze/sec-agent/agent/internal/config"
	"github.com/visiblaze/sec-agent/agent/internal/logging"
)

type Client struct {
	cfg    *config.Config
	logger *logging.Logger
	http   *http.Client
}

func NewClient(cfg *config.Config, logger *logging.Logger) *Client {
	return &Client{
		cfg:    cfg,
		logger: logger,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SendPayload(payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("%s/ingest", c.cfg.APIBaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.cfg.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
