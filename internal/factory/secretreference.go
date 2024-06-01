package factory

import (
	"context"
	"errors"
	"net/url"

	"token-manager/internal/secretreference/gitlab"

	"token-manager/internal/secretreference"
	"token-manager/internal/secretreference/gsm"
	"token-manager/internal/secretreference/onepassword"
	"token-manager/internal/secretreference/ssm"
)

var factoryMethods = map[string]func(context.Context, *url.URL) (secretreference.SecretReference, error){
	"op":     onepassword.NewFromURL,
	"gsm":    gsm.NewFromURL,
	"arn":    ssm.NewFromURL,
	"ssm":    ssm.NewFromURL,
	"gitlab": gitlab.NewFromURL,
}

var (
	SupportedScheme        = make([]string, 0, len(factoryMethods))
	UnsupportedSchemeError = errors.New("Unsupported scheme")
)

func NewSecretReferenceFromURL(ctx context.Context, referenceURL string) (secretreference.SecretReference, error) {
	var parsedURL *url.URL
	parsedURL, err := url.Parse(referenceURL)
	if err != nil {
		return nil, err
	}
	newFromURL, ok := factoryMethods[parsedURL.Scheme]
	if !ok {
		return nil, UnsupportedSchemeError
	}

	return newFromURL(ctx, parsedURL)
}

func init() {
	for scheme := range factoryMethods {
		SupportedScheme = append(SupportedScheme, scheme)
	}
}
