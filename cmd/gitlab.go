package cmd

import (
	"errors"
	"net/url"

	"token-manager/internal/secretreference"

	"github.com/spf13/cobra"
)

func NewGitlabCmdGroup(root *cobra.Command) *cobra.Command {
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
		if url, err := url.Parse(baseURL); err != nil || (url.Scheme != "https" || url.Host == "") {
			return errors.New("a valid https base url must be provided")
		}

		tokenUrl, err := cmd.Flags().GetString("admin-token-url")
		if err != nil {
			return err
		}
		if tokenUrl != "" {
			_, err = secretreference.NewFromURL(cmd.Context(), tokenUrl)
			if err != nil {
				return errors.New("admin-token-url is not a valid secret reference url")
			}
		}

		return nil
	}

	c.PersistentFlags().SortFlags = false
	c.PersistentFlags().String("url", "https://gitlab.com", "to rotate the token from")
	c.PersistentFlags().String("admin-token-url", "", "the URL to the secret containing the admin token")
	rootCmd.AddCommand(&c)
	return &c
}

// gitlabRootCmd represents the base command when called without any subcommands
var gitlabRootCmd = NewGitlabCmdGroup(rootCmd)
