package cmd

import (
	"log"
	"time"

	"token-manager/internal/factory"

	"github.com/spf13/cobra"

	"token-manager/internal/gitlab"
)

import "C"
import (
	"errors"
)

type gitlabCreateCommand struct {
	cobra.Command
	createToken gitlab.CreateTokenCommand
}

func newCreateCommand() *gitlabCreateCommand {
	c := &gitlabCreateCommand{
		Command: cobra.Command{
			Use:   "create token-url",
			Short: "create a group or project token and store it in the secret store",
			Args:  cobra.ExactArgs(1),
			Long:  `creates a new Gitlab token saves it in the secret store`,
		},
	}

	c.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		var err error

		if c.Parent() != nil && c.Parent().PersistentPreRunE != nil {
			err = c.Parent().PersistentPreRunE(cmd, args)
		}
		return err
	}

	c.PreRunE = func(cmd *cobra.Command, args []string) error {
		var err error
		var durationInDays int
		durationInDays, err = cmd.Flags().GetInt("duration")
		if err != nil || durationInDays < 1 || durationInDays > 365 {
			return errors.New("duration in days must be between 1 and 365")
		}
		c.createToken.Duration = time.Duration(durationInDays) * 24 * time.Hour
		if c.createToken.Url, err = cmd.Flags().GetString("url"); err != nil {
			return err
		}

		if c.createToken.Project != "" && c.createToken.Group != "" {
			return errors.New("--project and --project cannot be used together")
		}

		c.createToken.Token, err = factory.NewSecretReferenceFromURL(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		return nil
	}

	c.RunE = func(cmd *cobra.Command, args []string) error {
		err := c.createToken.Create(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}
		return nil
	}
	c.Flags().SortFlags = false
	c.Flags().IntVarP(&c.createToken.DurationInDays, "duration", "d", 30, "of the validity of the new token in days")
	c.Flags().StringVarP(&c.createToken.Project, "project", "p", "", "name of the gitlab project the token belongs to")
	c.Flags().StringVarP(&c.createToken.Group, "group", "g", "", "name of the gitlab group the token belongs to")
	c.Flags().StringVarP(&c.createToken.Name, "name", "n", "", "name of the gitlab token to create")
	c.Flags().StringSliceVarP(&c.createToken.Scopes, "scope", "s", []string{"read_repository"}, "scopes for the token, see https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#personal-access-token-scopes")
	c.Flags().VarP(&c.createToken.AccessLevel, "access-level", "a", "of the token: guest, reporter, developer, maintainer, owner")

	c.MarkFlagRequired("name")
	c.MarkFlagRequired("access-level")
	return c
}
