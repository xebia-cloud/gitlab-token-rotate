package onepassword

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/dvcrn/go-1password-cli/op"
)

type TokenReference struct {
	vaultName string
	itemName  string
	client    *op.Client
}

// NewTokenStore create a new 1password token reference.
func NewTokenReference(vaultName, itemName string) (*TokenReference, error) {
	if _, err := exec.LookPath("op"); err != nil {
		return nil, errors.New("op not found in $PATH")
	}

	return &TokenReference{vaultName: vaultName, itemName: itemName, client: op.NewOpClient()}, nil
}

// ReadToken reads the token from an API_CREDENTIAL in 1Password from the specified item and vault.
func (t TokenReference) ReadToken() (string, error) {
	item, err := t.client.VaultItem(t.itemName, t.vaultName)
	if err != nil {
		return "", err
	}
	if item.Category != "API_CREDENTIAL" {
		return "", errors.New("item found in vault is not of type API_CREDENTIAL")
	}

	for _, field := range item.Fields {
		if field.Label == "credential" {
			return field.Value, nil
		}
	}
	return "", errors.New("no credential found in item")
}

// UpdateToken updates the credential and expires field values of the specified item and vault.
func (t TokenReference) UpdateToken(token string, expiresAt time.Time) error {
	_, err := op.NewOpClient().EditItemField(t.vaultName, t.itemName,
		op.Assignment{Name: "credential", Value: token},
		op.Assignment{Name: "expires", Value: fmt.Sprintf("%d", expiresAt.Unix())},
	)
	return err
}
