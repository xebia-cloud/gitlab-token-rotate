package secretreference

import (
	"context"
	"fmt"
	"net/url"

	"token-manager/internal/secretreference/gsm"
	"token-manager/internal/secretreference/onepassword"
	"token-manager/internal/secretreference/ssm"
)

func NewFromURL(ctx context.Context, referenceURL string) (SecretReference, error) {
	var parsedURL *url.URL
	parsedURL, err := url.Parse(referenceURL)
	if err != nil {
		return nil, err
	}
	switch parsedURL.Scheme {
	case "op":
		return onepassword.NewFromURL(ctx, parsedURL)
	case "gsm":
		return gsm.NewFromURL(ctx, parsedURL)
	case "ssm":
		return ssm.NewFromURL(ctx, parsedURL)
	case "arn":
		return ssm.NewFromURL(ctx, parsedURL)
	default:
		return nil, fmt.Errorf("unsupported scheme %s", parsedURL.Scheme)
	}
}
