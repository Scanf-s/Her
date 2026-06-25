package server

import (
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/Scanf-s/her/internal/session"
	"github.com/Scanf-s/her/internal/understander"
)

// Handler serves WebSocket(ws) and runs conversation session for user.
type Handler struct {
	u understander.Understander
}

// NewHandler returns a Handler whose sessions are powered by understander.
func NewHandler(u understander.Understander) *Handler {
	return &Handler{u: u}
}

type clientMsg struct {
	// Text type user input
	Text string `json:"text"`

	// Audio type user input
	Audio []byte `json:"audio"`
}

type serverMsg struct {
	Type string `json:"type"` // token, done, error
	Text string `json:"text,omitempty"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP protocol to WS
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(16 * 1024 * 1024) // 16MB read limit for audio input

	// Initialize session
	ctx := r.Context()
	sess := session.New(h.u)

	for {
		// Receive user message through websocket
		var input clientMsg
		if err := wsjson.Read(ctx, conn, &input); err != nil {
			// if client closed or context canceled
			return
		}

		// Send user text (session will stream the reply token from the LLM interface)
		tokens, err := sess.Turn(ctx, understander.UserTurn{Text: input.Text, Audio: input.Audio})
		if err != nil {
			_ = wsjson.Write(ctx, conn, serverMsg{Type: "error", Text: err.Error()})
			continue
		}
		for tok := range tokens {
			// Pass the message to the WebSocket upon receiving it.
			if err := wsjson.Write(ctx, conn, serverMsg{Type: "token", Text: tok}); err != nil {
				return
			}
		}
		if err := wsjson.Write(ctx, conn, serverMsg{Type: "done"}); err != nil {
			return
		}
	}
}
