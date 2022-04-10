package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// repository:ekristen/docker-cache:push
//  - type: repository
//  - subj: ekristen/docker-cache
//  - acts: push
// registry:catalog:*
//  - type: registry
//  - subj: catalog
//  - acts: *

func TestParseScope(t *testing.T) {
	cases := []struct {
		Scope  string
		Scopes []Scope
	}{
		{
			Scope: "registry:catalog:*",
			Scopes: []Scope{
				{
					Type:    "registry",
					Name:    "catalog",
					Actions: []string{"*"},
				},
			},
		},
		{
			Scope: "repository:ekristen/dockit:pull",
			Scopes: []Scope{
				{
					Type:    "repository",
					Name:    "ekristen/dockit",
					Actions: []string{"pull"},
				},
			},
		},
		{
			Scope: "repository:ekristen/dockit:pull,push",
			Scopes: []Scope{
				{
					Type:    "repository",
					Name:    "ekristen/dockit",
					Actions: []string{"pull", "push"},
				},
			},
		},
		{
			Scope: "repository:ekristen/dockit:pull repository:ekristen/dockit2:pull",
			Scopes: []Scope{
				{
					Type:    "repository",
					Name:    "ekristen/dockit",
					Actions: []string{"pull"},
				},
				{
					Type:    "repository",
					Name:    "ekristen/dockit2",
					Actions: []string{"pull"},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Scope, func(t *testing.T) {
			scopes, err := ParseScope(c.Scope)
			assert.NoError(t, err)
			assert.Equal(t, c.Scopes, scopes)
		})
	}
}

func TestParseScopeFailure(t *testing.T) {
	cases := []struct {
		Scope string
	}{
		{
			Scope: "repository:ekristen/dockit:pull:hi",
		},
		{
			Scope: "repository:ekristen/dockit",
		},
		{
			Scope: "repository:ekristen",
		},
	}

	for _, c := range cases {
		t.Run(c.Scope, func(t *testing.T) {
			_, err := ParseScope(c.Scope)
			assert.Error(t, err)
		})
	}
}
