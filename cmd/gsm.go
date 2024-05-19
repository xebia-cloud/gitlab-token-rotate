package cmd

/*
Copyright 2024 Xebia Nederland B.V.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"context"
	"net/url"

	"github.com/spf13/cobra"
	"gitlab-token-rotate/internal"
	"gitlab-token-rotate/internal/token/gsm"
)

// googleSecretManagerCmd rotates the token in a Google Secret Manager secret.
var googleSecretManagerCmd = &cobra.Command{
	Use:   "gsm",
	Short: "rotate the token in a  Google Secret Manager secret",
	Long:  `rotates the Gitlab token in Google Secret Manager`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := func() (*url.URL, error) {
			u, _ := cmd.Flags().GetString("url")
			return url.Parse(u)
		}()
		secretName, _ := cmd.Flags().GetString("secret-name")
		project, _ := cmd.Flags().GetString("project")
		useDefaultCredentials, _ := cmd.Flags().GetBool("use-default-credentials")
		durationInDays, _ := cmd.Flags().GetInt("duration")
		ctx := context.Background()

		reference, err := gsm.NewTokenReference(ctx, secretName, project, useDefaultCredentials)
		if err != nil {
			return err
		}

		rotator := internal.NewRotator(durationInDays, url, true)
		return rotator.Rotate(ctx, reference)
	},
}

func init() {
	rootCmd.AddCommand(googleSecretManagerCmd)
	googleSecretManagerCmd.Flags().String("secret-name", "", "of the secret in which the token is stored")
	googleSecretManagerCmd.Flags().String("project", "", "in which the secret is stored")
	googleSecretManagerCmd.Flags().Bool("use-default-credentials", false, "to authenticate with")
	googleSecretManagerCmd.MarkFlagRequired("name")
}
