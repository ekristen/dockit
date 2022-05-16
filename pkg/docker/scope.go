package docker

import (
	"fmt"
	"regexp"
	"strings"
)

var typeRegexp = regexp.MustCompile(`^([a-z0-9]+)(\([a-z0-9]+\))?$`)

type Scope struct {
	Type    string   `json:"type"`
	Class   string   `json:"class,omitempty"`
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

func ParseScope(raw string) (scopes []Scope, err error) {
	rawRequests := strings.Split(raw, " ")

	for _, r := range rawRequests {
		parts := strings.Split(r, ":")

		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid scope: %s", r)
		}

		resourceType, resourceName, actions := parts[0], parts[1], parts[2]

		resourceType, resourceClass := splitResourceClass(resourceType)
		if resourceType == "" {
			continue
		}

		scopes = append(scopes, Scope{
			Type:    resourceType,
			Class:   resourceClass,
			Name:    resourceName,
			Actions: strings.Split(actions, ","),
		})
	}

	return scopes, err
}

func splitResourceClass(t string) (string, string) {
	matches := typeRegexp.FindStringSubmatch(t)
	if len(matches) < 2 {
		return "", ""
	}
	if len(matches) == 2 || len(matches[2]) < 2 {
		return matches[1], ""
	}
	return matches[1], matches[2][1 : len(matches[2])-1]
}
