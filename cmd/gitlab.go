package cmd

import (
	"errors"
	"net/url"

	"github.com/spf13/cobra"
)

func newGitlabCmdGroup() *cobra.Command {
	c := cobra.Command{
		Use:   "gitlab",
		Short: "Manage your gitlab tokens",
		Long:  `create and rotate Gitlab tokens`,
	}

	c.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if c.Parent() != nil && c.Parent().PersistentPreRunE != nil {
			if err := c.Parent().PersistentPreRunE(cmd, args); err != nil {
				return err
			}
		}

		baseURL, err := cmd.Flags().GetString("url")
		if err != nil || baseURL == "" {
			return errors.New("base url must be provided")
		}
		ServerUrl, err := url.Parse(baseURL)
		if err != nil || (ServerUrl.Scheme != "https" || ServerUrl.Host == "") {
			return errors.New("a valid https base url must be provided")
		}

		return nil
	}

	c.PersistentFlags().SortFlags = false
	c.PersistentFlags().String("url", "https://gitlab.com", "to rotate the token from")

	c.AddCommand(&newRotateCommand().Command)
	return &c
}
