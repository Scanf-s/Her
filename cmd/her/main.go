package main

import (
	"log"
	"net/http"

	"github.com/Scanf-s/her/internal/config"
	"github.com/Scanf-s/her/internal/server"
	"github.com/Scanf-s/her/internal/understander"
)

func main() {
	cfg := config.Load()
	u := understander.NewLiteRTLM(cfg.LLMEndpoint, cfg.LLMModel, cfg.SystemPrompt)

	mux := http.NewServeMux()
	mux.Handle("/ws", server.NewHandler(u))
	mux.Handle("/", http.FileServer(http.Dir("web")))

	log.Printf("Web server is running on %s and She is listening on %s, model %s)", cfg.Addr, cfg.LLMEndpoint, cfg.LLMModel)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Fatal(err)
	}
}
