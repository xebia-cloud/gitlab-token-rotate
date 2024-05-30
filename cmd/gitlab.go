package cmd

import (
	"errors"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"token-manager/internal/factory"
	"token-manager/internal/gitlab"
)

func NewGitlabCmdGroup() *cobra.Command {
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

		var token string
		tokenUrl, err := cmd.Flags().GetString("admin-token-url")
		if err != nil {
			return err
		}
		if tokenUrl != "" {
			tokenReference, err := factory.NewSecretReferenceFromURL(cmd.Context(), tokenUrl)
			if err != nil {
				return errors.New("admin-token-url is not a valid secret reference url")
			}

			token, err = tokenReference.Read(cmd.Context())
			if err != nil {
				return err
			}
		} else {
			token = os.Getenv("GITLAB_TOKEN")
		}
		// initialize global admin client
		if _, err = gitlab.GetAdminClient(cmd.Context(), ServerUrl.Host, token); err != nil {
			return err
		}

		return nil
	}

	c.PersistentFlags().SortFlags = false
	c.PersistentFlags().String("url", "https://gitlab.com", "to rotate the token from")
	c.PersistentFlags().String("admin-token-url", "", "the URL to the secret containing the admin token")

	c.AddCommand(&NewRotateCommand().Command)
	c.AddCommand(NewReadCmd())
	return &c
}
