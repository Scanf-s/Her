package config

import (
	"fmt"
	"os"
)

// Config holds all runtime settings for the server.
type Config struct {
	// HTTP listen address, e.g. ":8080"
	Addr string

	// base URL of an OpenAI-compatible server (litert-lm serve address)
	LLMEndpoint string

	// model name to request
	LLMModel string

	// instructions for the conversation
	SystemPrompt string
}

const defaultSystemPrompt = "Your name is Samantha, a warm and patient English conversation partner. " +
	"Keep the conversation natural and flowing, and ask friendly follow-up questions. " +
	"Do not correct the user's grammar or word choice during the conversation. " +
	"Keep your replies short but natural(1 to 3 sentences). so the user does most of the talking."

// Load reads configuration from the environment, applying defaults for any
// unset or empty variable.
func Load() Config {
	return Config{
		Addr:        getenv("HER_ADDR", ":8080"),
		LLMEndpoint: getenv("HER_LLM_ENDPOINT", "http://localhost:8081"),
		LLMModel:    getenv("HER_LLM_MODEL", "gemma4-e4b"),
		SystemPrompt: getenv(
			"HER_SYSTEM_PROMPT",
			defaultSystemPrompt+
				fmt.Sprintf("Just chat naturally within the CEFR English level %s. ",
					getenv("HER_LLM_ENGLISH_LEVEL", "1"),
				),
		),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
