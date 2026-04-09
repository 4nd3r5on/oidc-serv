package session

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an active user login session.
// The ID is an opaque high-entropy token handed to the client
// (cookie or Authorization header); everything else stays server-side.
type Session struct {
	Key       string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

// Data is the auth context payload for MethodSession requests.
type Data struct {
	UserID     uuid.UUID
	SessionKey string
}
