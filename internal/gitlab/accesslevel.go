package gitlab

import (
	"fmt"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type AccessLevel struct {
	value gitlab.AccessLevelValue
}

var accessLevels = map[string]AccessLevel{
	"no":         {gitlab.NoPermissions},
	"minimal":    {gitlab.MinimalAccessPermissions},
	"guest":      {gitlab.GuestPermissions},
	"reporter":   {gitlab.ReporterPermissions},
	"developer":  {gitlab.DeveloperPermissions},
	"maintainer": {gitlab.MaintainerPermissions},
	"owner":      {gitlab.OwnerPermissions},
	"admin":      {gitlab.AdminPermissions},
}

var invertedAccessLevels = make(map[AccessLevel]string)

func init() {
	for name, value := range accessLevels {
		invertedAccessLevels[value] = name
	}
}

func CreateAccessLevelFromString(s string) (AccessLevel, error) {
	if level, ok := accessLevels[strings.ToLower(s)]; ok {
		return level, nil
	}
	return AccessLevel{gitlab.NoPermissions}, fmt.Errorf("invalid access level: %s", s)
}

func (a *AccessLevel) String() string {
	if result, ok := invertedAccessLevels[*a]; ok {
		return result
	} else {
		return fmt.Sprintf("%d", *a)
	}
}

func (a *AccessLevel) Set(value string) (err error) {
	*a, err = CreateAccessLevelFromString(value)
	return err
}

func (a *AccessLevel) Type() string {
	return "AccessLevel"
}
