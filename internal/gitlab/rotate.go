package gitlab

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/xanzy/go-gitlab"

	"token-manager/internal/secretreference"
)

type GitlabRotateCommand struct {
	Url        string
	Token      secretreference.SecretReference
	AdminToken secretreference.SecretReference
	Project    string
	Group      string
	Duration   time.Duration
}

func (c GitlabRotateCommand) Rotate(ctx context.Context) error {
	var err error
	var token string
	var tokenClient, adminClient *gitlab.Client

	token, err = c.Token.Read(ctx)
	if err != nil {
		return err
	}

	tokenClient, err = gitlab.NewClient(token, gitlab.WithBaseURL(c.Url))
	if err != nil {
		return err
	}

	adminClient, err = gitlab.NewClient(os.Getenv("GITLAB_TOKEN"), gitlab.WithBaseURL(c.Url))
	if err != nil {
		return err
	}

	accessToken, _, err := tokenClient.PersonalAccessTokens.GetSinglePersonalAccessToken()
	if err != nil {
		return err
	}

	if slices.Index(accessToken.Scopes, "api") == -1 {
		fmt.Printf("accessToken %s does not have the permission to rotate itself", accessToken.Name)
	} else {
		adminClient = tokenClient
	}

	log.Printf("api token %s will expire on %s", accessToken.Name, accessToken.ExpiresAt.String())

	var tokenName, newToken string
	var newExpirationDate time.Time

	if c.Project != "" {
		tokenName, newToken, newExpirationDate, err = c.rotateProjectAccessToken(adminClient, accessToken.ID)
	} else if c.Group != "" {
		tokenName, newToken, newExpirationDate, err = c.rotateGroupAccessToken(adminClient, accessToken.ID)
	} else {
		tokenName, newToken, newExpirationDate, err = c.rotatePersonalAccessToken(adminClient, accessToken.ID)
	}

	if err != nil {
		return err
	}

	err = c.Token.Update(ctx, newToken, newExpirationDate)
	if err != nil {
		log.Printf("Error updating the gitlab access token in 1password. Manual renewal and update to %s is required",
			c.Token)
		writeTokenToTemporaryFile(newToken)
		return err
	}

	log.Printf("rotated api token %s, will expire on %s",
		tokenName, newExpirationDate.Format(time.DateOnly))

	return nil
}

func (c GitlabRotateCommand) ExpirationDate() *gitlab.ISOTime {
	newExpirationDate := time.Now().Add(c.Duration).Truncate(time.Hour * 24)
	return (*gitlab.ISOTime)(&newExpirationDate)
}

func (c GitlabRotateCommand) rotatePersonalAccessToken(client *gitlab.Client, tokenID int) (string, string, time.Time, error) {
	newAccessToken, _, err := client.PersonalAccessTokens.RotatePersonalAccessToken(
		tokenID,
		&gitlab.RotatePersonalAccessTokenOptions{
			ExpiresAt: c.ExpirationDate(),
		})
	if err != nil {
		return "", "", time.Time{}, err
	}
	return newAccessToken.Name, newAccessToken.Token, time.Time(*newAccessToken.ExpiresAt), nil
}

func (c GitlabRotateCommand) rotateProjectAccessToken(client *gitlab.Client, tokenID int) (string, string, time.Time, error) {
	newAccessToken, _, err := client.ProjectAccessTokens.RotateProjectAccessToken(
		c.Project,
		tokenID,
		&gitlab.RotateProjectAccessTokenOptions{
			ExpiresAt: c.ExpirationDate(),
		})
	if err != nil {
		return "", "", time.Time{}, err
	}
	return newAccessToken.Name, newAccessToken.Token, time.Time(*newAccessToken.ExpiresAt), nil
}

func (c GitlabRotateCommand) rotateGroupAccessToken(client *gitlab.Client, tokenID int) (string, string, time.Time, error) {
	newAccessToken, _, err := client.GroupAccessTokens.RotateGroupAccessToken(
		c.Group,
		tokenID,
		&gitlab.RotateGroupAccessTokenOptions{
			ExpiresAt: c.ExpirationDate(),
		})
	if err != nil {
		return "", "", time.Time{}, err
	}
	return newAccessToken.Name, newAccessToken.Token, time.Time(*newAccessToken.ExpiresAt), nil
}

// writeTokenToTemporaryFile writes token to temporary file to recover from a failed attempt to
// write to the token store.
func writeTokenToTemporaryFile(token string) {
	file, err := os.CreateTemp("/tmp", "gl-token-")
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
