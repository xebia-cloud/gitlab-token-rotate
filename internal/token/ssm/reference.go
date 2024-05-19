package ssm

import (
	"context"
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

// NewTokenReference create a new AWS SSM parameter token reference.
func NewTokenReference(ctx context.Context, parameterName string, awsRegion string) (*TokenReference, error) {
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

// ReadToken reads the token from the SSM parameter
func (t TokenReference) ReadToken(ctx context.Context) (string, error) {
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
func (t TokenReference) UpdateToken(ctx context.Context, token string, expiresAt time.Time) error {
	_, err := t.client.PutParameter(ctx,
		&awsssm.PutParameterInput{
			Name:      aws.String(t.parameterName),
			Value:     aws.String(token),
			Overwrite: aws.Bool(true),
		})

	return err
}
