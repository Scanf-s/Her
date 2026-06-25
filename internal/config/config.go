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
	"First, react to what the user said and how they feel; then ask one natural follow-up question about it. " +
	"Do not correct the user's grammar or word choice during the conversation. Just do chat. " +
	"Use clear, friendly language a learner can easily follow. " +
	"Keep your replies short but natural (1-3 sentences) so the user does most of the talking."

const englishAssessmentSystemPrompt = "You are a supportive English tutor reviewing a learner's messages from a finished conversation. " +
	"Look only at the user's messages (ignore the AI assistant's). Find unnatural phrasing, grammar mistakes, and awkward word choices. " +
	"Choose the 3-5 most useful issues. For each, output one bullet in this exact format:\n" +
	"- \"<what the user wrote>\" -> \"<a more natural version>\" (short reason)\n" +
	"End with one encouraging sentence. If the English is already good, say so briefly instead of inventing issues."

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
