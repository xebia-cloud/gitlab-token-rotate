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

type gitlabRotateCommand struct {
	cobra.Command
	gitlabRotate gitlab.GitlabRotateCommand
}

func newRotateCommand() *gitlabRotateCommand {
	c := &gitlabRotateCommand{
		Command: cobra.Command{
			Use:   "rotate token-url",
			Short: "rotate the token stored in a secret store",
			Args:  cobra.ExactArgs(1),
			Long:  `reads the Gitlab token from the secret store and rotates it`,
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
		c.gitlabRotate.Duration = time.Duration(durationInDays) * 24 * time.Hour
		if c.gitlabRotate.Url, err = cmd.Flags().GetString("url"); err != nil {
			return err
		}
		if c.gitlabRotate.Project, err = cmd.Flags().GetString("project"); err != nil {
			return err
		}
		if c.gitlabRotate.Group, err = cmd.Flags().GetString("group"); err != nil {
			return err
		}
		if c.gitlabRotate.Project != "" && c.gitlabRotate.Group != "" {
			return errors.New("--project and --project cannot be used together")
		}

		if adminToken := cmd.Flag("admin-token-url").Value.String(); adminToken != "" {
			c.gitlabRotate.AdminToken, err = factory.NewSecretReferenceFromURL(cmd.Context(), adminToken)
		}

		c.gitlabRotate.Token, err = factory.NewSecretReferenceFromURL(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		return nil
	}

	c.RunE = func(cmd *cobra.Command, args []string) error {
		err := c.gitlabRotate.Rotate(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}
		return nil
	}
	c.Flags().Int("duration", 30, "of the validity of the rotated token in days")
	c.Flags().SortFlags = false
	c.Flags().String("project", "", "name of the gitlab project the token belongs to")
	c.Flags().String("group", "", "name of the gitlab group the token belongs to")
	return c
}
