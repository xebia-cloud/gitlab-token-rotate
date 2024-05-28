package ssm

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
)

type TokenReference struct {
	parameterName string
	awsRegion     string
	client        *awsssm.Client
}

func (t TokenReference) String() string {
	if strings.HasPrefix(t.parameterName, "arn:") {
		return t.parameterName
	}
	return fmt.Sprintf("ssm://%s", t.parameterName)
}

// NewTokenReference create a new AWS SSM parameter token reference.
func NewTokenReference(ctx context.Context, parameterName string, awsRegion string) (*TokenReference, error) {
	if !strings.HasPrefix(parameterName, "arn:") && !strings.HasPrefix(parameterName, "/") {
		parameterName = "/" + parameterName
	}

	ref := TokenReference{
		parameterName: parameterName, awsRegion: awsRegion,
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	ref.client = awsssm.NewFromConfig(cfg, func(o *awsssm.Options) {
		if awsRegion != "" {
			o.Region = awsRegion
		}
	})
	return &ref, nil
}

var parameterPattern = regexp.MustCompile(`^(?P<Partition>[^:]*):ssm:(?P<Region>[^:]*):(?P<AccountID>[^:]*):parameter/(?P<Resource>.*)$`)

func NewFromURL(ctx context.Context, referenceURL *url.URL) (*TokenReference, error) {
	if referenceURL.Scheme == "arn" {
		if !parameterPattern.MatchString(referenceURL.Opaque) {
			return nil, fmt.Errorf("unsupported ARN %s", referenceURL.Scheme)
		}

		return NewTokenReference(ctx, "arn:"+referenceURL.Opaque, "")
	}

	if referenceURL.Scheme == "ssm" {
		return NewTokenReference(ctx, referenceURL.Path, "")
	}

	return nil, fmt.Errorf("unsupported URL %s", referenceURL)
}

// ReadToken reads the token from the SSM parameter
func (t TokenReference) Read(ctx context.Context) (string, error) {
	response, err := t.client.GetParameter(ctx,
		&awsssm.GetParameterInput{
			Name:           aws.String(t.parameterName),
			WithDecryption: aws.Bool(true),
		})
	if err != nil {
		return "", err
	}
	return *response.Parameter.Value, nil
}

// UpdateToken updates the SSM parameter with the token
func (t TokenReference) Update(ctx context.Context, token string, expiresAt time.Time) error {
	_, err := t.client.PutParameter(ctx,
		&awsssm.PutParameterInput{
			Name:      aws.String(t.parameterName),
			Value:     aws.String(token),
			Overwrite: aws.Bool(true),
		})

	return err
}
