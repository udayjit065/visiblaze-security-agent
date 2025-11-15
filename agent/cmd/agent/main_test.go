package main

import "testing"

func TestVersionExists(t *testing.T) {
	// Test that Version constant is defined and not empty
	if Version == "" {
		t.Fatal("Version should not be empty")
	}
	if Version != "0.1.0" {
		t.Logf("Current version: %s", Version)
	}
}
