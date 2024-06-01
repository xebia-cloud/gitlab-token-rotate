package env

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"

	"token-manager/internal/factory"
	"token-manager/internal/secretreference"
)

// resolve resolves environment variable secret references to their value
func resolve(ctx context.Context, env []string, factoryMethod func(context.Context, string) (secretreference.SecretReference, error)) (map[string]string, error) {
	result := make(map[string]string)
	for _, variable := range env {
		parts := strings.SplitN(variable, "=", 2)
		if len(parts) < 2 {
			parts = append(parts, "")
		}
		name := parts[0]
		referenceURL := parts[1]
		_, err := url.Parse(referenceURL)
		if err != nil {
			continue
		}
		secretReference, err := factoryMethod(ctx, referenceURL)
		if errors.Is(err, factory.UnsupportedSchemeError) {
			continue
		}
		if err != nil {
			return nil, err
		}
		value, err := secretReference.Read(ctx)
		if err != nil {
			return nil, err
		}
		result[name] = value
	}
	return result, nil
}

// UpdateEnvironment updates the environment variables secret references with the actual value.
func UpdateEnvironment(ctx context.Context) error {
	variables, err := resolve(ctx, os.Environ(), factory.NewSecretReferenceFromURL)
	if err != nil {
		return err
	}
	for key, value := range variables {
		err = os.Setenv(key, value)
		if err != nil {
			return err
		}
	}
	return nil
}
