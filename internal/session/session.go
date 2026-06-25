package session

import (
	"context"
	"strings"
	"sync"

	"github.com/Scanf-s/her/internal/understander"
)

// Session is one ongoing conversation.
// Safe for sequential use.
// A caller must fully drain the channel returned by Turn before calling Turn again.
type Session struct {
	u       understander.Understander
	mu      sync.Mutex
	history []understander.Message
}

// New creates a Session backed by understander.
func New(u understander.Understander) *Session {
	return &Session{u: u}
}

// Turn sends one user message and returns a channel that streams the assistant reply tokens.
// When the channel closes, both the user message and the full assistant reply will be appended to the conversation history.
func (s *Session) Turn(ctx context.Context, userInput understander.UserTurn) (<-chan string, error) {

	// A history should be locked by mutex.
	// Because of parallel write request on below goroutine to the history list.
	s.mu.Lock()
	hist := append([]understander.Message(nil), s.history...)
	s.mu.Unlock()

	// Request to the LLM with user's message
	// Reply message from LLM will be streamed through channel
	// Audio type takes precedence over Text type
	reply, err := s.u.Respond(ctx, hist, userInput)
	if err != nil {
		return nil, err
	}

	// Create channel to send received tokens (words: string type) to caller.
	// unbuffered channel (cuz the goroutine has to wait until next token received)
	// And in unbuffered channel, out has been blocked until someone receives the data
	out := make(chan string)
	go func() {
		defer close(out)
		var b strings.Builder
		// from streaming reply from LLM, hand over to the Turn caller
		for tok := range reply.Tokens {
			b.WriteString(tok)
			select {
			case out <- tok:
			case <-ctx.Done():
			}
		}

		// Mutex lock for writing on the history array
		s.mu.Lock()
		s.history = append(s.history,
			understander.Message{Role: understander.RoleUser, Content: reply.UserTranscript},
			understander.Message{Role: understander.RoleAssistant, Content: b.String()},
		)
		s.mu.Unlock()
	}()
	return out, nil
}
