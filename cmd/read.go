package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"token-manager/internal/secretreference"
)

func NewReadCmd(root *cobra.Command) *cobra.Command {
	c := &cobra.Command{
		Use:   "read token-url",
		Short: "Read a secret from the secret store",
		Args:  cobra.MinimumNArgs(1),
	}

	c.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if c.Parent() != nil && c.Parent().PersistentPreRunE != nil {
			if err := c.Parent().PersistentPreRunE(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}

	c.RunE = func(cmd *cobra.Command, args []string) error {
		tokenReference, err := secretreference.NewFromURL(cmd.Context(), args[0])
		if err != nil {
			log.Fatal(err)
		}
		token, err := tokenReference.Read(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}
		_, err = fmt.Printf("%s", token)
		return err
	}

	rootCmd.AddCommand(c)
	return c
}

func init() {
	NewReadCmd(rootCmd)
}
