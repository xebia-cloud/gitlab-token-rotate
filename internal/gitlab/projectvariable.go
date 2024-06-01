package gitlab

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	gl "github.com/xanzy/go-gitlab"
)

type ProjectTokenReference struct {
	url              *url.URL
	project          string
	key              string
	environmentScope string
}

func (t ProjectTokenReference) EnvironmentScope() string {
	if t.environmentScope == "" {
		return "*"
	}
	return t.environmentScope
}

func (t ProjectTokenReference) String() string {
	return t.url.String()
}

func (t ProjectTokenReference) Scheme() string {
	return t.url.Scheme
}

// isASingleVariable checks of the reference points to a single CI/CD project variable
func (t ProjectTokenReference) isASingleVariable(client *gl.Client) error {
	if t.environmentScope == "" {
		variables, _, err := client.ProjectVariables.ListVariables(t.project, &gl.ListProjectVariablesOptions{})
		if err != nil {
			return err
		}

		environments := make([]string, 0, len(variables))
		for _, variable := range variables {
			if variable.Key == t.key {
				environments = append(environments, variable.EnvironmentScope)
			}
		}
		if len(environments) > 1 {
			return fmt.Errorf("no environment scope was specified, but variable has multiple scopes: %s",
				strings.Join(environments, ", "))
		}
	}
	return nil
}

// Read reads the token from the gitlab project CI/CD variable
func (t ProjectTokenReference) Read(ctx context.Context) (token string, err error) {
	var client *gl.Client
	client, err = gl.NewClient(os.Getenv("GITLAB_TOKEN"))
	if err != nil {
		return "", err
	}

	if err = t.isASingleVariable(client); err != nil {
		return "", err
	}

	variable, _, err := client.ProjectVariables.GetVariable(
		t.project,
		t.key,
		&gl.GetProjectVariableOptions{
			&gl.VariableFilter{t.EnvironmentScope()},
		},
	)
	if err != nil {
		return "", err
	}
	return variable.Value, nil
}

// Update updates the token in the the gitlab project CI/CD variable
func (t ProjectTokenReference) Update(ctx context.Context, token string, expiresAt time.Time) (err error) {
	var client *gl.Client
	client, err = gl.NewClient(os.Getenv("GITLAB_TOKEN"))
	if err != nil {
		return err
	}

	if err = t.isASingleVariable(client); err != nil {
		return err
	}

	environmentScope := t.EnvironmentScope()
	_, _, err = client.ProjectVariables.UpdateVariable(t.project, t.key,
		&gl.UpdateProjectVariableOptions{Value: &token, EnvironmentScope: &environmentScope})
	return err
}
