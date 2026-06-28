package session_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Scanf-s/her/internal/session"
	"github.com/Scanf-s/her/internal/understander"
)

// stubUnderstander records the history it was called with and streams a fixed
// list of tokens.
type stubUnderstander struct {
	tokens     []string
	lastCallHi []understander.Message
}

func (s *stubUnderstander) Respond(ctx context.Context, history []understander.Message, turn understander.UserTurn) (understander.Reply, error) {
	s.lastCallHi = history
	ch := make(chan string)
	toks := s.tokens
	go func() {
		defer close(ch)
		for _, tok := range toks {
			ch <- tok
		}
	}()
	return understander.Reply{UserTranscript: turn.Text, Tokens: ch}, nil
}

func TestSession_Turn_streamsTokens(t *testing.T) {
	stub := &stubUnderstander{tokens: []string{"He", "llo"}}
	s := session.New(stub)

	tokens, err := s.Turn(context.Background(), understander.UserTurn{Text: "hi"})
	if err != nil {
		t.Fatalf("Turn: %v", err)
	}
	var got strings.Builder
	for tok := range tokens {
		got.WriteString(tok)
	}
	if got.String() != "Hello" {
		t.Errorf("reply = %q, want %q", got.String(), "Hello")
	}
}

func TestSession_Turn_recordsHistory(t *testing.T) {
	stub := &stubUnderstander{tokens: []string{"Hello"}}
	s := session.New(stub)

	first, _ := s.Turn(context.Background(), understander.UserTurn{Text: "hi"})
	for range first { // drain so the assistant message is recorded
	}

	second, _ := s.Turn(context.Background(), understander.UserTurn{Text: "bye"})
	for range second {
	}

	// On the second call the stub should have seen the first exchange.
	if len(stub.lastCallHi) != 2 {
		t.Fatalf("history length = %d, want 2", len(stub.lastCallHi))
	}
	if stub.lastCallHi[0].Role != understander.RoleUser || stub.lastCallHi[0].Content != "hi" {
		t.Errorf("history[0] = %+v, want user/hi", stub.lastCallHi[0])
	}
	if stub.lastCallHi[1].Role != understander.RoleAssistant || stub.lastCallHi[1].Content != "Hello" {
		t.Errorf("history[1] = %+v, want assistant/Hello", stub.lastCallHi[1])
	}
}
