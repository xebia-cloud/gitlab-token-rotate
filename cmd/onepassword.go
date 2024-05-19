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
	"errors"
	"net/url"

	"github.com/spf13/cobra"
	"gitlab-token-rotate/internal"
	"gitlab-token-rotate/internal/token/onepassword"
)

// onePasswordCmd rotates the token in 1 password
var onePasswordCmd = &cobra.Command{
	Use:   "1password",
	Short: "rotate the token stored in 1password",
	Long:  `reads the Gitlab token from 1password and stores the new token with the new expiration date`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		duration, err := cmd.Flags().GetInt("duration")
		if err != nil || duration <= 0 || duration > 365 {
			return errors.New("token duration must be between 1 and 365 days")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := url.Parse(cmd.Flag("url").Value.String())
		vaultName := cmd.Flag("vault").Value.String()
		itemName := cmd.Flag("item").Value.String()
		durationInDays, _ := cmd.Flags().GetInt("duration")

		reference, err := onepassword.NewTokenReference(vaultName, itemName)
		if err != nil {
			return err
		}

		rotator := internal.NewRotator(durationInDays, url)
		return rotator.Rotate(reference)
	},
}

func init() {
	rootCmd.AddCommand(onePasswordCmd)
	onePasswordCmd.Flags().String("vault", "Private", "the token is stored in")
	onePasswordCmd.MarkFlagRequired("vault")
	onePasswordCmd.Flags().String("item", "", "of the token")
	onePasswordCmd.MarkFlagRequired("item")
}
