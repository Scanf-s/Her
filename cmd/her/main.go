package main

import (
	"github.com/Scanf-s/her/internal/config"
	"github.com/Scanf-s/her/internal/understander"
)

func main() {
	cfg := config.Load()
	u := understander.NewLiteRTLM(cfg.LLMEndpoint, cfg.LLMModel, cfg.SystemPrompt)
}
