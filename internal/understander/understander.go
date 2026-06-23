// Package understander turns a user's turn (text now, audio later) into a streamed assistant reply.
package understander

import "context"

// Role values for a conversation Message.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Message is one entry in the conversation history.
type Message struct {
	// RoleSystem | RoleUser | RoleAssistant
	Role string

	Content string
}

// UserTurn is a single input from the user.
type UserTurn struct {
	// Text is for the text-based chat.
	Text string

	// Audio is for the audio-native chat.
	Audio []byte
}

// Reply is an assistant's response to a UserTurn.
type Reply struct {

	// UserTranscript is what the user said.
	UserTranscript string

	// Tokens streams the assistant reply.
	// The channel is closed when the reply is complete.
	// Callers MUST drain it (defer).
	Tokens <-chan string
}

// Understander produces a streamed reply for one user turn, given prior history.
// Implementations must be safe for sequential use by a single Session.
// The reason why I choose a streamed reply is for the real-time, fastest respond to a user.
// ctx: Context
// history: Message history slice
// turn: User's input
// it returns Reply struct
type Understander interface {
	Respond(ctx context.Context, history []Message, turn UserTurn) (Reply, error)
}
