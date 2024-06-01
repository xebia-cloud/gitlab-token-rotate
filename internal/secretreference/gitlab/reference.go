package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"token-manager/internal/secretreference"
)

var (
	urlRegex    = regexp.MustCompile("^/(projects|groups)/(.*)/variables/([^/]+)$")
	expectError = errors.New("expected gitlab://<host>/(projects|groups)/variables/<name>")
)

// NewFromURL create a new Gitlab variable reference
func NewFromURL(ctx context.Context, referenceUrl *url.URL) (secretreference.SecretReference, error) {
	var err error
	var id string
	var key string

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
	if id, err = url.QueryUnescape(match[2]); err != nil {
		return nil, expectError
	}

	if key, err = url.QueryUnescape(match[3]); err != nil {
		return nil, expectError
	}

	q, err := url.ParseQuery(referenceUrl.RawQuery)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(referenceUrl.Path, "/groups/") {
		if len(q) > 0 {
			return nil, fmt.Errorf("query parameters are not support for group variable references")
		}

		return &GroupTokenReference{
			url:   referenceUrl,
			group: id,
			key:   key,
		}, nil
	}

	if len(q) > 1 {
		return nil, fmt.Errorf("multiple query parameters are not support for project variable references")
	}
	if environment, ok := q["environment"]; ok && len(environment) > 1 {
		return nil, expectError
	}
	return &ProjectTokenReference{
		url:              referenceUrl,
		project:          id,
		key:              key,
		environmentScope: q.Get("environment"),
	}, nil
}
