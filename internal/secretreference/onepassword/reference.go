package onepassword

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/dvcrn/go-1password-cli/op"
)

type TokenReference struct {
	vaultName string
	itemName  string
	client    *op.Client
}

func (t TokenReference) String() string {
	return fmt.Sprintf("op://%s/%s", t.vaultName, t.itemName)
}

func NewFromURL(ctx context.Context, referenceURL *url.URL) (*TokenReference, error) {
	// op://Private/gitlab access token/note
	if referenceURL.Scheme != "op" {
		return nil, errors.New("unsupported schema " + referenceURL.Scheme + ":")
	}
	paths := strings.Split(referenceURL.Path[1:], "/")
	var vaultName, itemName string

	vaultName = referenceURL.Host
	if vaultName != "" && len(paths) == 1 {
		itemName = paths[0]
	} else if vaultName == "" && len(paths) == 2 {
		vaultName = paths[0]
		itemName = paths[1]
	} else {
		return nil, errors.New("expected an url in the form op://<vault>/<item name or id>")
	}
	return NewTokenReference(ctx, vaultName, itemName)
}

// NewTokenStore create a new 1password token reference.
func NewTokenReference(_ context.Context, vaultName, itemName string) (*TokenReference, error) {
	if _, err := exec.LookPath("op"); err != nil {
		return nil, errors.New("op not found in $PATH")
	}

	return &TokenReference{vaultName: vaultName, itemName: itemName, client: op.NewOpClient()}, nil
}

// ReadToken reads the token from an API_CREDENTIAL in 1Password from the specified item and vault.
func (t TokenReference) Read(_ context.Context) (string, error) {
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
func (t TokenReference) Update(_ context.Context, token string, expiresAt time.Time) error {
	_, err := op.NewOpClient().EditItemField(t.vaultName, t.itemName,
		op.Assignment{Name: "credential", Value: token},
		op.Assignment{Name: "expires", Value: fmt.Sprintf("%d", expiresAt.Unix())},
	)
	return err
}
