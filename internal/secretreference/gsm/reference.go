package gsm

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"

	"github.com/binxio/gcloudconfig"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type TokenReference struct {
	secretName            string
	secretVersion         string
	project               string
	useDefaultCredentials bool
	client                *secretmanager.Client
}

func (t TokenReference) String() string {
	return fmt.Sprintf("gsm:///%s", t.secretName)
}

// NewTokenReference create a new Google Secret Manager token reference
func NewTokenReference(ctx context.Context, secretName string, project string, useDefaultCredentials bool) (*TokenReference, error) {
	var err error
	var credentials *google.Credentials

	if useDefaultCredentials || !gcloudconfig.IsGCloudOnPath() {
		credentials, err = google.FindDefaultCredentials(ctx)
	} else {
		credentials, err = gcloudconfig.GetCredentials("")
	}
	if err != nil {
		return nil, err
	}

	if project == "" {
		project = credentials.ProjectID
	}
	if project == "" {
		return nil, fmt.Errorf("no google project defined")
	}

	var ref TokenReference
	ref.client, err = secretmanager.NewClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, err
	}

	ref.secretName = secretName
	ref.secretVersion, err = normalizeSecretName(secretName, project)
	if err != nil {
		return nil, err
	}

	return &ref, nil
}

func NewFromURL(ctx context.Context, referenceURL *url.URL) (*TokenReference, error) {
	if referenceURL.Scheme != "gsm" {
		return nil, fmt.Errorf("unsupported scheme %s", referenceURL.Scheme)
	}
	if referenceURL.Host != "" {
		return nil, errors.New("expected an url in the form gsm:///<google secret name>")
	}
	return NewTokenReference(ctx, referenceURL.Path[1:], "", false)
}

// ReadToken reads the token from an Google Secret Manager secret
func (t TokenReference) Read(ctx context.Context) (string, error) {
	request := &secretmanagerpb.AccessSecretVersionRequest{
		Name: t.secretVersion,
	}

	response, err := t.client.AccessSecretVersion(ctx, request)
	if err != nil {
		return "", err
	}

	return string(response.Payload.Data), nil
}

// UpdateToken updates the secret of the Google Secret Manager secret
func (t TokenReference) Update(ctx context.Context, token string, expiresAt time.Time) error {
	parent := t.secretVersion[:strings.Index(t.secretVersion, "/versions/")]

	request := &secretmanagerpb.AddSecretVersionRequest{
		Parent:  parent,
		Payload: &secretmanagerpb.SecretPayload{Data: []byte(token)},
	}

	_, err := t.client.AddSecretVersion(ctx, request)
	return err
}

// normalizeSecretName normalizes the Google Secret Manager secret name to "projects/[^/]+/secrets/[^/]+/.*"
func normalizeSecretName(secretName string, project string) (string, error) {
	var name string
	var version string

	if match, _ := regexp.MatchString("projects/[^/]+/secrets/[^/]+/.*", secretName); match {
		return secretName, nil
	}

	parts := strings.Split(secretName, "/")
	switch len(parts) {
	case 1:
		name = parts[0]
		version = "latest"
	case 2:
		if match, _ := regexp.MatchString("([0-9]+|latest)", parts[1]); match {
			name = parts[0]
			version = parts[1]
		} else {
			project = parts[0]
			name = parts[1]
			version = "latest"
		}
	case 3:
		project = parts[0]
		name = parts[1]
		version = parts[2]
	default:
		return secretName, fmt.Errorf("invalid secret name specification: %s", secretName)
	}
	result := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, name, version)
	return result, nil
}
