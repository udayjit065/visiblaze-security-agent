package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Logger struct {
	logDir string
	file   *os.File
}

type logEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Error     string      `json:"error,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func New(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return nil, err
	}

	logPath := fmt.Sprintf("%s/agent.log", logDir)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return nil, err
	}

	return &Logger{logDir: logDir, file: file}, nil
}

func (l *Logger) log(level, msg string, err error, data interface{}) {
	entry := logEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Message:   msg,
		Data:      data,
	}
	if err != nil {
		entry.Error = err.Error()
	}

	jsonBytes, _ := json.Marshal(entry)
	fmt.Fprintln(l.file, string(jsonBytes))
}

func (l *Logger) Infof(msg string, args ...interface{}) {
	l.log("INFO", fmt.Sprintf(msg, args...), nil, nil)
}

func (l *Logger) Errorf(msg string, args ...interface{}) {
	l.log("ERROR", fmt.Sprintf(msg, args...), nil, nil)
}

func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.log("WARN", fmt.Sprintf(msg, args...), nil, nil)
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
