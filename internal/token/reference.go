package token

import (
	"context"
	"time"
)

// Reference references a token in a token store.
type Reference interface {
	ReadToken(ctx context.Context) (string, error)
	UpdateToken(ctx context.Context, token string, expiresAt time.Time) error
}
