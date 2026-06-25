package understander

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// LiteRTLM is an Understander backed by an OpenAI-compatible streaming endpoint
type LiteRTLM struct {
	// LiteRTLM API endpoint
	endpoint string

	// AI backend model (now, Gemma4-E4B)
	model string

	// system prompt
	system string

	// client for http request
	client *http.Client
}

// NewLiteRTLM returns a LiteRTLM struct that talks to an OpenAI-compatible server.
// It requests to the given model.
// if non-empty system(system prompt) provided, is prepended as a system message on every request.
func NewLiteRTLM(endpoint, model, system string) *LiteRTLM {
	return &LiteRTLM{endpoint: endpoint, model: model, system: system, client: http.DefaultClient}
}

// chatRequest DTO
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// chatMessage DTO is following https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create spec
type chatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// contentPart DTO is following https://developers.openai.com/api/reference/python/resources/chat/subresources/completions/methods/create
type contentPart struct {
	Type       string      `json:"type"`
	InputAudio *inputAudio `json:"input_audio,omitempty"`
}

// inputAudio DTO is following https://developers.openai.com/api/reference/python/resources/chat/subresources/completions/methods/create#(resource)%20chat.completions%20%3E%20(model)%20chat_completion_content_part_input_audio%20%3E%20(schema)%20%3E%20(property)%20input_audio
type inputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"` // wav, mp3 (wav preferred in Gemma4)
}

// streamChunk follows the OpenAI Streaming event spec
// https://developers.openai.com/api/reference/resources/chat/subresources/completions/streaming-events
type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// Respond implements Understander.
func (l *LiteRTLM) Respond(ctx context.Context, history []Message, turn UserTurn) (Reply, error) {

	// reserve message slice
	messages := make([]chatMessage, 0, len(history)+2)

	// If system prompt provided, append to the message slice
	if l.system != "" {
		messages = append(messages, chatMessage{Role: RoleSystem, Content: l.system})
	}

	// Append previous messages
	for _, m := range history {
		messages = append(messages, chatMessage{Role: m.Role, Content: m.Content})
	}

	// Append user provided message
	if turn.Audio != nil {
		messages = append(messages, chatMessage{Role: RoleUser, Content: []contentPart{
			{
				Type: "input_audio",
				InputAudio: &inputAudio{
					Data:   base64.StdEncoding.EncodeToString(turn.Audio),
					Format: "wav",
				},
			},
		}})
	} else {
		messages = append(messages, chatMessage{Role: RoleUser, Content: turn.Text})
	}

	// HTTP request setup
	body, err := json.Marshal(
		chatRequest{Model: l.model, Messages: messages, Stream: true},
	)
	if err != nil {
		// failed to serialize request DTO
		return Reply{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.endpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		// failed to initialize request object
		return Reply{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request to LLM
	resp, err := l.client.Do(req)
	if err != nil {
		// got an error from LLM
		return Reply{}, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return Reply{}, fmt.Errorf("unexpected http status %s from LLM", resp.Status)
	}

	// token streaming
	tokens := make(chan string)
	// goroutine will receive a token streaming in the background.
	// Another request will not be blocked to wait the next token from the LLM
	go func() {
		// close response body and token receiver channel if LLM stops to stream
		defer resp.Body.Close()
		defer close(tokens)

		sc := bufio.NewScanner(resp.Body)
		// receive 64KB per request streaming to ensure the long respond text line
		// maximum bound of buffer availability: 1MB
		sc.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), 1024*1024)

		// for loop waits until the new next stream received
		for sc.Scan() {
			line := sc.Text()
			data, ok := strings.CutPrefix(line, "data: ")
			if !ok {
				continue
			}

			if data == "[DONE]" {
				// finish the goroutine
				return
			}

			// Send LLM text respond through channel (tokens will receive this)
			var chunk streamChunk
			if json.Unmarshal([]byte(data), &chunk) != nil {
				continue
			}
			for _, c := range chunk.Choices {
				if c.Delta.Content == "" {
					continue
				}
				select {
				case tokens <- c.Delta.Content:
				// context timeout
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Stream the token respond
	transcript := turn.Text
	if turn.Audio != nil {
		// Will be implemented after adding Kokoro interface (TTS)
		transcript = "[Audio Message]"
	}
	return Reply{UserTranscript: transcript, Tokens: tokens}, nil
}
