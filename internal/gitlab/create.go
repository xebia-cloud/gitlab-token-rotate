package gitlab

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xanzy/go-gitlab"

	"token-manager/internal/secretreference"
)

type CreateTokenCommand struct {
	Url            string
	Token          secretreference.SecretReference
	AccessLevel    AccessLevel
	Scopes         []string
	Project        string
	Group          string
	Duration       time.Duration
	DurationInDays int
	Name           string
}

func (c CreateTokenCommand) Create(ctx context.Context) error {
	var err error
	var token string
	var adminClient *gitlab.Client

	if _, err = c.Token.Read(ctx); err != nil {
		return fmt.Errorf("The secret to store the token in, does not exist or cannot be read, %s", err)
	}

	adminClient, err = gitlab.NewClient(os.Getenv("GITLAB_TOKEN"), gitlab.WithBaseURL(c.Url))
	if err != nil {
		return err
	}

	if c.Project != "" {
		if projectAccessTokens, _, listErr := adminClient.ProjectAccessTokens.ListProjectAccessTokens(c.Project, &gitlab.ListProjectAccessTokensOptions{
			Page:    0,
			PerPage: 100,
		}); listErr == nil {
			for _, accessToken := range projectAccessTokens {
				if accessToken.Name == c.Name {
					return errors.New("An access token with the same name already exists")
				}
			}
		}

		var projectToken *gitlab.ProjectAccessToken
		projectToken, _, err = adminClient.ProjectAccessTokens.CreateProjectAccessToken(
			c.Project, &gitlab.CreateProjectAccessTokenOptions{
				Name:        &c.Name,
				ExpiresAt:   c.ExpirationDate(),
				Scopes:      &c.Scopes,
				AccessLevel: &c.AccessLevel.value,
			})
		if err != nil {
			return err
		}
		token = projectToken.Token
	} else if c.Group != "" {
		if groupAccessTokens, _, listErr := adminClient.GroupAccessTokens.ListGroupAccessTokens(c.Group, &gitlab.ListGroupAccessTokensOptions{
			Page:    0,
			PerPage: 100,
		}); listErr == nil {
			for _, accessToken := range groupAccessTokens {
				if accessToken.Name == c.Name {
					return errors.New("An access token with the same name already exists")
				}
			}
		}

		var groupToken *gitlab.GroupAccessToken
		groupToken, _, err = adminClient.GroupAccessTokens.CreateGroupAccessToken(
			c.Project, &gitlab.CreateGroupAccessTokenOptions{
				Name:        &c.Name,
				ExpiresAt:   c.ExpirationDate(),
				Scopes:      &c.Scopes,
				AccessLevel: &c.AccessLevel.value,
			})
		if err != nil {
			return err
		}
		token = groupToken.Token
	} else {
		return errors.New("personal access token cannot be created using the API")
	}

	err = c.Token.Update(ctx, token, time.Time(*c.ExpirationDate()))
	if err != nil {
		log.Printf("Error storing the gitlab access token. Manual renewal and update to %s is required",
			c.Token)
		writeTokenToTemporaryFile(token)
		return err
	}

	log.Printf("new api token %s, will expire on %s",
		c.Name, time.Time(*c.ExpirationDate()).Format(time.DateOnly))

	return nil
}

func (c CreateTokenCommand) ExpirationDate() *gitlab.ISOTime {
	newExpirationDate := time.Now().Add(c.Duration).Truncate(time.Hour * 24)
	return (*gitlab.ISOTime)(&newExpirationDate)
}
