package gitlab

import (
	"context"
	"net/url"
	"os"
	"time"

	gl "github.com/xanzy/go-gitlab"
)

type GroupTokenReference struct {
	url   *url.URL
	group string
	key   string
}

func (t GroupTokenReference) String() string {
	return t.url.String()
}

func (t GroupTokenReference) Scheme() string {
	return t.url.Scheme
}

// Read the token from the gitlab group CI/CD variable
func (t GroupTokenReference) Read(ctx context.Context) (token string, err error) {
	var client *gl.Client
	client, err = gl.NewClient(os.Getenv("GITLAB_TOKEN"))
	if err != nil {
		return "", err
	}
	variable, _, err := client.GroupVariables.GetVariable(
		t.group,
		t.key)
	if err != nil {
		return "", err
	}
	return variable.Value, nil
}

// Update  the token in the gitlab group CI/CD variable
func (t GroupTokenReference) Update(ctx context.Context, token string, expiresAt time.Time) (err error) {
	var client *gl.Client
	client, err = gl.NewClient(os.Getenv("GITLAB_TOKEN"))
	if err != nil {
		return err
	}

	_, _, err = client.GroupVariables.UpdateVariable(t.group, t.key,
		&gl.UpdateGroupVariableOptions{Value: &token})
	return err
}
