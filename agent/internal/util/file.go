package util

import (
	"os"
	"strings"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ReadFileLines(path string) ([]string, error) {
	content, err := ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(content), "\n"), nil
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func EnsureFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
