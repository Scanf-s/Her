package config_test

import (
	"testing"

	"github.com/Scanf-s/her/internal/config"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("HER_ADDR", "")
	t.Setenv("HER_LLM_ENDPOINT", "")
	cfg := config.Load()
	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":8080")
	}
	if cfg.LLMEndpoint != "http://localhost:8081" {
		t.Errorf("LLMEndpoint = %q, want %q", cfg.LLMEndpoint, "http://localhost:8081")
	}
	if cfg.SystemPrompt == "" {
		t.Error("SystemPrompt should have a non-empty default")
	}
}

func TestLoad_envOverride(t *testing.T) {
	t.Setenv("HER_ADDR", ":9999")
	cfg := config.Load()
	if cfg.Addr != ":9999" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":9999")
	}
}
