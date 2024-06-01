package cmd

import (
	"fmt"
	"log"

	"token-manager/internal/factory"

	"github.com/spf13/cobra"
)

func newReadCmd() *cobra.Command {
	c := new(cobra.Command)
	c.Use = "read token-url"
	c.Short = "Read a secret from the secret store"
	c.Args = cobra.MinimumNArgs(1)

	c.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if c.Parent() != nil && c.Parent().PersistentPreRunE != nil {
			if err := c.Parent().PersistentPreRunE(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}

	c.RunE = func(cmd *cobra.Command, args []string) error {
		tokenReference, err := factory.NewSecretReferenceFromURL(cmd.Context(), args[0])
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

	return c
}
