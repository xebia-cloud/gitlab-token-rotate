package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	gl "github.com/xanzy/go-gitlab"
)

type TokenReference struct {
	url              *url.URL
	id               string
	key              string
	environmentScope string
}

func (t TokenReference) String() string {
	return fmt.Sprintf("gitlab://%s/%s", t.url.Host, t.url.Path)
}

var (
	urlRegex    = regexp.MustCompile("^/(projects|groups)/(.*)/variables/([^/]+)$")
	expectError = errors.New("expected gitlab://<host>/(projects|groups)/variables/<name>?environment=<scope>")
)

// NewTokenReference create a new Gitlab token reference
func NewFromURL(ctx context.Context, referenceUrl *url.URL) (*TokenReference, error) {
	var err error

	t := TokenReference{url: referenceUrl}

	if referenceUrl.Scheme != "gitlab" {
		return nil, fmt.Errorf("unsupported scheme %s", referenceUrl.Scheme)
	}

	if referenceUrl.Host == "" {
		return nil, expectError
	}

	match := urlRegex.FindStringSubmatch(referenceUrl.Path)
	if match == nil && len(match) != 4 {
		return nil, expectError
	}
	if t.id, err = url.QueryUnescape(match[2]); err != nil {
		return nil, expectError
	}

	if t.key, err = url.QueryUnescape(match[3]); err != nil {
		return nil, expectError
	}

	q, err := url.ParseQuery(referenceUrl.RawQuery)
	if err != nil {
		return nil, err
	}

	if len(q) > 1 {
		return nil, expectError
	}
	if environment, ok := q["environment"]; ok && len(environment) > 1 {
		return nil, expectError
	}
	t.environmentScope = q.Get("environment")

	return &t, nil
}

// ReadToken reads the token from an Google Secret Manager secret
func (t TokenReference) Read(ctx context.Context) (token string, err error) {
	var client *gl.Client
	client, err = GetAdminClient(ctx, t.url.Host, "")
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(t.url.Path, "/groups") {
		variable, _, err := client.GroupVariables.GetVariable(
			t.id,
			t.key)
		if err != nil {
			return "", err
		}
		return variable.Value, nil
	}

	if t.environmentScope == "" {
		variables, _, err := client.ProjectVariables.ListVariables(t.id, &gl.ListProjectVariablesOptions{})
		if err != nil {
			return "", err
		}

		environments := make([]string, 0, len(variables))
		for _, variable := range variables {
			if variable.Key == t.key {
				environments = append(environments, variable.EnvironmentScope)
			}
		}
		if len(environments) > 1 {
			return "", fmt.Errorf("no environment scope was specified, but variable has multiple scopes: %s",
				strings.Join(environments, ", "))
		}
	}
	variable, _, err := client.ProjectVariables.GetVariable(
		t.id,
		t.key, &gl.GetProjectVariableOptions{&gl.VariableFilter{t.environmentScope}},
	)
	if err != nil {
		return "", err
	}

	return variable.Value, err
}

// UpdateToken updates the token in the the gitlab CI/CD variable
func (t TokenReference) Update(ctx context.Context, token string, expiresAt time.Time) error {
	if strings.HasPrefix(t.url.Path, "/groups") {
	}
	return errors.New("todo")
}
