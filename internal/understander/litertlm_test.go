package understander_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Scanf-s/her/internal/understander"
)

func TestLiteRTLM_Respond_streamsTokens(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hi\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\" there\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	l := understander.NewLiteRTLM(srv.URL, "test-model", "be nice")
	reply, err := l.Respond(context.Background(), nil, understander.UserTurn{Text: "hello"})
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}

	var got strings.Builder
	for tok := range reply.Tokens {
		got.WriteString(tok)
	}
	if got.String() != "Hi there" {
		t.Errorf("tokens = %q, want %q", got.String(), "Hi there")
	}
	if reply.UserTranscript != "hello" {
		t.Errorf("UserTranscript = %q, want %q", reply.UserTranscript, "hello")
	}
}

func TestLiteRTLM_Respond_errorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	l := understander.NewLiteRTLM(srv.URL, "test-model", "")
	_, err := l.Respond(context.Background(), nil, understander.UserTurn{Text: "hi"})
	if err == nil {
		t.Fatal("expected error on 500 status, got nil")
	}
}
