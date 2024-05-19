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

	"gitlab-token-rotate/internal/token/ssm"

	"github.com/spf13/cobra"
	"gitlab-token-rotate/internal"
)

// awsSsmCmd rotates the token in a AWS SSM parameter
var awsSsmCmd = &cobra.Command{
	Use:   "aws-ssm",
	Short: "rotate the token in an AWS SSM parameter",
	Long:  `rotates the Gitlab token in an AWS SSM parameter`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		durationInDays, _ := cmd.Flags().GetInt("duration")
		url, _ := func() (*url.URL, error) {
			u, _ := cmd.Flags().GetString("url")
			return url.Parse(u)
		}()
		name, _ := cmd.Flags().GetString("parameter-name")
		region, _ := cmd.Flags().GetString("region")
		ctx := context.Background()

		reference, err := ssm.NewTokenReference(ctx, name, region)
		if err != nil {
			return err
		}

		rotator := internal.NewRotator(durationInDays, url, true)
		return rotator.Rotate(ctx, reference)
	},
}

func init() {
	rootCmd.AddCommand(awsSsmCmd)
	awsSsmCmd.Flags().String("name", "", "of the SSM Parameter")
	awsSsmCmd.Flags().String("region", "", "in which the parameter lives")
	awsSsmCmd.MarkFlagRequired("name")
}
