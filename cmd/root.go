package cmd

import (
	"os"
	"token-manager/internal/env"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "token-manager",
	Short: "Create and rotate tokens",
	Long:  `Create and rotate tokens in a secret store`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return env.UpdateEnvironment(cmd.Context())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	rootCmd.AddCommand(newReadCmd())
	rootCmd.AddCommand(newGitlabCmdGroup())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
