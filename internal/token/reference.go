package token

import "time"

// Reference references a token in a token store.
type Reference interface {
	ReadToken() (string, error)
	UpdateToken(token string, expiresAt time.Time) error
}
