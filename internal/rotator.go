package internal

import (
	"context"
	"errors"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/xanzy/go-gitlab"
	"gitlab-token-rotate/internal/token"
)

type Rotator struct {
	durationInDays int
	baseURL        *url.URL
	rescueToken    bool
}

// NewRotator creates a new gitlab personal access token rotator
func NewRotator(durationInDays int, baseURL *url.URL, rescueToken bool) Rotator {
	return Rotator{durationInDays: durationInDays, baseURL: baseURL, rescueToken: rescueToken}
}

// Rotate rotates the gitlab personal access token from the token store.
func (r Rotator) Rotate(ctx context.Context, tokenReference token.Reference) error {
	if r.durationInDays < 1 || r.durationInDays > 365 {
		return errors.New("rotator: duration in days must be between 1 and 365")
	}

	token, err := tokenReference.ReadToken(ctx)
	if err != nil {
		return err
	}

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(r.baseURL.String()))
	if err != nil {
		return err
	}

	accessToken, _, err := client.PersonalAccessTokens.GetSinglePersonalAccessToken()
	if err != nil {
		return err
	}

	log.Printf("api token %s will expire on %s", accessToken.Name, accessToken.ExpiresAt.String())

	newExpirationDate := time.Now().Add(time.Hour * 24 * time.Duration(r.durationInDays))
	newAccessToken, _, err := client.PersonalAccessTokens.RotatePersonalAccessToken(
		accessToken.ID,
		&gitlab.RotatePersonalAccessTokenOptions{
			ExpiresAt: (*gitlab.ISOTime)(&newExpirationDate),
		})
	if err != nil {
		return err
	}

	err = tokenReference.UpdateToken(ctx, newAccessToken.Token, time.Time(*newAccessToken.ExpiresAt))
	if err != nil {
		log.Printf("Error updating the gitlab access token in 1password. Manual renewal is required")
		if r.rescueToken {
			writeTokenToTemporaryFile(newAccessToken.Token)
		}
		return err
	}

	log.Printf("rotated api token %s, will expire on %s",
		newAccessToken.Name, newAccessToken.ExpiresAt.String())

	return nil
}

// writeTokenToTemporaryFile writes token to temporary file to recover from a failed attempt to
// write to the token store.
func writeTokenToTemporaryFile(token string) {
	file, err := os.CreateTemp("/tmp", "token-")
	if err != nil {
		log.Printf("failed to create temporary file to store token, %s", err.Error())
		return
	}
	defer file.Close()
	if err = os.Chmod(file.Name(), 0o600); err != nil {
		log.Printf("failed to chmod %s, %s", file.Name(), err.Error())
	}

	if err = os.WriteFile(file.Name(), []byte(token), 0o600); err != nil {
		log.Printf("failed to write token to temporary file, %s", err.Error())
		return
	}
	log.Printf("token written to %s", file.Name())
}
