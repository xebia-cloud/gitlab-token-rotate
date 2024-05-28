package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"slices"
	"time"

	"token-manager/internal/secretreference"

	"github.com/xanzy/go-gitlab"
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

func rotatePersonalAccessToken(client *gitlab.Client, tokenID int, newExpirationDate time.Time) (string, string, time.Time, error) {
	newAccessToken, _, err := client.PersonalAccessTokens.RotatePersonalAccessToken(
		tokenID,
		&gitlab.RotatePersonalAccessTokenOptions{
			ExpiresAt: (*gitlab.ISOTime)(&newExpirationDate),
		})
	if err != nil {
		return "", "", time.Time{}, err
	}
	return newAccessToken.Name, newAccessToken.Token, time.Time(*newAccessToken.ExpiresAt), nil
}

func rotateProjectAccessToken(client *gitlab.Client, project string, tokenID int, newExpirationDate time.Time) (string, string, time.Time, error) {
	newAccessToken, _, err := client.ProjectAccessTokens.RotateProjectAccessToken(
		project,
		tokenID,
		&gitlab.RotateProjectAccessTokenOptions{
			ExpiresAt: (*gitlab.ISOTime)(&newExpirationDate),
		})
	if err != nil {
		return "", "", time.Time{}, err
	}
	return newAccessToken.Name, newAccessToken.Token, time.Time(*newAccessToken.ExpiresAt), nil
}

func rotateGroupAccessToken(client *gitlab.Client, group string, tokenID int, newExpirationDate time.Time) (string, string, time.Time, error) {
	newAccessToken, _, err := client.GroupAccessTokens.RotateGroupAccessToken(
		group,
		tokenID,
		&gitlab.RotateGroupAccessTokenOptions{
			ExpiresAt: (*gitlab.ISOTime)(&newExpirationDate),
		})
	if err != nil {
		return "", "", time.Time{}, err
	}
	return newAccessToken.Name, newAccessToken.Token, time.Time(*newAccessToken.ExpiresAt), nil
}

// Rotate rotates the gitlab personal access token from the token store.
func (r Rotator) Rotate(ctx context.Context, project, group string, tokenReference, adminTokenReference secretreference.SecretReference) error {
	var err error
	var tokenClient, adminClient *gitlab.Client
	if r.durationInDays < 1 || r.durationInDays > 365 {
		return errors.New("rotator: duration in days must be between 1 and 365")
	}

	token, err := tokenReference.Read(ctx)
	if err != nil {
		return err
	}

	tokenClient, err = gitlab.NewClient(token, gitlab.WithBaseURL(r.baseURL.String()))
	if err != nil {
		return err
	}

	if adminTokenReference != nil {
		adminToken, err := adminTokenReference.Read(ctx)
		if err != nil {
			return err
		}

		adminClient, err = gitlab.NewClient(adminToken, gitlab.WithBaseURL(r.baseURL.String()))
		if err != nil {
			return err
		}
	} else {
		adminClient = tokenClient
	}

	accessToken, _, err := tokenClient.PersonalAccessTokens.GetSinglePersonalAccessToken()
	if err != nil {
		return err
	}

	if adminTokenReference == nil && slices.Index(accessToken.Scopes, "api") == -1 {
		return fmt.Errorf("rotator: accessToken %s does not have the permission to rotate itself", accessToken.Name)
	}

	log.Printf("api token %s will expire on %s", accessToken.Name, accessToken.ExpiresAt.String())

	newExpirationDate := time.Now().Add(time.Hour * 24 * time.Duration(r.durationInDays))

	var tokenName, newToken string

	if project != "" {
		tokenName, newToken, newExpirationDate, err = rotateProjectAccessToken(adminClient, project, accessToken.ID, newExpirationDate)
	} else if group != "" {
		tokenName, newToken, newExpirationDate, err = rotateGroupAccessToken(adminClient, group, accessToken.ID, newExpirationDate)
	} else {
		tokenName, newToken, newExpirationDate, err = rotatePersonalAccessToken(adminClient, accessToken.ID, newExpirationDate)
	}

	if err != nil {
		return err
	}

	err = tokenReference.Update(ctx, newToken, newExpirationDate)
	if err != nil {
		log.Printf("Error updating the gitlab access token in 1password. Manual renewal is required")
		if r.rescueToken {
			writeTokenToTemporaryFile(newToken)
		}
		return err
	}

	log.Printf("rotated api token %s, will expire on %s",
		tokenName, newExpirationDate.Format(time.DateOnly))

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
