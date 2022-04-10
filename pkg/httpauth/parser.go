// Package auth provides functions for extracting a user Auth token from a
// request and associating it with a Context.
package httpauth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

type Auth interface {
	Schema() string
	Token() string
	Username() string
	Password() string
}

type Basic struct {
	username string
	password string
}

func (b *Basic) Schema() string {
	return BASIC_SCHEMA
}
func (b *Basic) Username() string {
	return b.username
}
func (b *Basic) Password() string {
	return b.password
}
func (b *Basic) Token() string {
	return b.password
}

type Bearer struct {
	token string
}

func (b *Bearer) Schema() string {
	return BEARER_SCHEMA
}
func (b *Bearer) Token() string {
	return b.token
}
func (b *Bearer) Username() string {
	return b.token
}
func (b *Bearer) Password() string {
	return b.token
}

const (
	BASIC_SCHEMA  string = "Basic "
	BEARER_SCHEMA string = "Bearer "
)

func Parse(req *http.Request) (a Auth, err error) {
	authHeader := req.Header.Get("authorization")
	if authHeader == "" {
		return a, errors.New("authorization header required")
	}

	if !strings.HasPrefix(authHeader, BASIC_SCHEMA) && !strings.HasPrefix(authHeader, BEARER_SCHEMA) {
		return a, errors.New("authorization requires Basic/Bearer scheme")
	}

	if strings.HasPrefix(authHeader, BASIC_SCHEMA) {
		str, err := base64.StdEncoding.DecodeString(authHeader[len(BASIC_SCHEMA):])
		if err != nil {
			return a, errors.New("base64 encoding issue")
		}
		b := &Basic{
			username: strings.Split(string(str), ":")[0],
			password: strings.Split(string(str), ":")[1],
		}
		return b, nil
	} else {
		b := &Bearer{
			token: authHeader[len(BEARER_SCHEMA):],
		}
		return b, nil
	}
}
