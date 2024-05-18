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
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/dvcrn/go-1password-cli/op"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

func getTokenFromVault(vaultName, itemName string) (*op.Item, error) {
	if _, err := exec.LookPath("op"); err != nil {
		return nil, errors.New("op not found in $PATH")
	}
	client := op.NewOpClient()

	item, err := client.VaultItem(itemName, vaultName)
	if err != nil {
		return nil, err
	}

	if item.Category != "API_CREDENTIAL" {
		return nil, errors.New("item found in vault is not of type API_CREDENTIAL")
	}

	return item, nil
}

func getApiToken(item *op.Item) (string, error) {
	for _, field := range item.Fields {
		if field.Label == "credential" {
			return field.Value, nil
		}
	}
	return "", errors.New("no credential found in item")
}

func getExpirationDate(item *op.Item) (time.Time, error) {
	for _, field := range item.Fields {
		if field.Label == "expires" && field.Type == "DATE" {
			expires, err := strconv.Atoi(field.Value)
			if err != nil {
				return time.Now(), err
			}
			return time.Unix(int64(expires), 0), nil
		}
	}
	return time.Now(), errors.New("no credential found in item")
}

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
		url := rootCmd.Flag("url").Value.String()
		vaultName := cmd.Flag("vault").Value.String()
		itemName := cmd.Flag("item").Value.String()
		durationInDays, _ := cmd.Flags().GetInt("duration")

		item, err := getTokenFromVault(vaultName, itemName)
		if err != nil {
			return err
		}
		token, err := getApiToken(item)
		if err != nil {
			return err
		}

		client, err := gitlab.NewClient(token, gitlab.WithBaseURL(url))
		if err != nil {
			return err
		}

		accessToken, _, err := client.PersonalAccessTokens.GetSinglePersonalAccessToken()
		if err != nil {
			return err
		}

		log.Printf("api token %s in %s will expire on %s",
			itemName, vaultName, accessToken.ExpiresAt)

		newExpirationDate := time.Now().Add(time.Hour * 24 * time.Duration(durationInDays))
		newAccessToken, _, err := client.PersonalAccessTokens.RotatePersonalAccessToken(
			accessToken.ID,
			&gitlab.RotatePersonalAccessTokenOptions{
				ExpiresAt: (*gitlab.ISOTime)(&newExpirationDate),
			})
		if err != nil {
			return err
		}

		newExpiresAt := fmt.Sprintf("%d", time.Time(*newAccessToken.ExpiresAt).Unix())
		item, err = op.NewOpClient().EditItemField(vaultName, itemName,
			op.Assignment{Name: "credential", Value: newAccessToken.Token},
			op.Assignment{Name: "expires", Value: newExpiresAt},
		)
		if err != nil {
			log.Printf("Error updating the gitlab access token in 1password. Manual renewal is required")
			return err
		}

		log.Printf("rotated api token %s in %s will expire on %s",
			itemName, vaultName, newAccessToken.ExpiresAt)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(onePasswordCmd)
	onePasswordCmd.Flags().String("vault", "Private", "the token is stored in")
	onePasswordCmd.MarkFlagRequired("vault")
	onePasswordCmd.Flags().String("item", "", "of the token")
	onePasswordCmd.MarkFlagRequired("item")
}
