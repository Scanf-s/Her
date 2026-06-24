package server_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/Scanf-s/her/internal/server"
	"github.com/Scanf-s/her/internal/understander"
)

type stubUnderstander struct{ tokens []string }

func (s *stubUnderstander) Respond(ctx context.Context, history []understander.Message, turn understander.UserTurn) (understander.Reply, error) {
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

func TestHandler_streamsReplyThenDone(t *testing.T) {
	h := server.NewHandler(&stubUnderstander{tokens: []string{"Hi", "!"}})
	srv := httptest.NewServer(h)
	defer srv.Close()

	ctx := context.Background()
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.CloseNow()

	if err := wsjson.Write(ctx, c, map[string]string{"text": "hello"}); err != nil {
		t.Fatalf("write: %v", err)
	}

	var got strings.Builder
	for {
		var msg struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := wsjson.Read(ctx, c, &msg); err != nil {
			t.Fatalf("read: %v", err)
		}
		switch msg.Type {
		case "token":
			got.WriteString(msg.Text)
		case "done":
			if got.String() != "Hi!" {
				t.Errorf("reply = %q, want %q", got.String(), "Hi!")
			}
			return
		default:
			t.Fatalf("unexpected message type %q", msg.Type)
		}
	}
}
