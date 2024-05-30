package factory

import (
	"context"
	"fmt"
	"log"
	"testing"
)

func Must[T any](obj T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return obj
}

func TestNewFromURL(t *testing.T) {
	type args struct {
		referenceURL string
	}
	ctx := context.Background()
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"aws ARN",
			args{"arn:aws:ssm:eu-central-1:12345435454456:parameter/gitlab/pat"},
			false,
		},
		{
			"onepassword URL",
			args{"op://Private/gitlab access token"},
			false,
		},
		{
			"Google Secret Manager URL",
			args{"gsm:///gitlab-pat"},
			false,
		},
		{
			"AWS Secret manager parameter",
			args{"ssm:///gitlab/pat"},
			false,
		},
		{
			"AWS Secret manager parameter ARN",
			args{"arn:aws:ssm:eu-central-1:123456789012:parameter/gitlab/pat"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSecretReferenceFromURL(ctx, tt.args.referenceURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSecretReferenceFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.referenceURL != fmt.Sprintf("%s", got) {
				t.Errorf("NewSecretReferenceFromURL() got = %v, want %v", got, tt.args)
			}
		})
	}
}
