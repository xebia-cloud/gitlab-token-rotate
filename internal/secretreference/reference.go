package secretreference

import (
	"context"
	"time"
)

// SecretReference references a token in a token store.
type SecretReference interface {
	Read(ctx context.Context) (string, error)
	Update(ctx context.Context, token string, expiresAt time.Time) error
}
