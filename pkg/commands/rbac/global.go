package rbac

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/docker/pkg/homedir"
	"github.com/urfave/cli/v2"
)

var (
	configDir     string
	homeDir       string
	configFileDir = ".docker"

	rbacFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "base-url",
			Value:   "http://localhost:4315/v2",
			EnvVars: []string{"DOCKIT_BASE_URL", "DOCKIT_GRANT_BASE_URL"},
		},
		&cli.StringFlag{
			Name:    "registry-url",
			EnvVars: []string{"DOCKIT_REGISTRY_URL", "DOCKIT_GRANT_REGISTRY_URL"},
		},
		&cli.BoolFlag{
			Name:    "insecure",
			EnvVars: []string{"DOCKIT_INSECURE", "DOCKIT_GRANT_INSECURE"},
			Value:   true,
		},
		&cli.StringFlag{
			Name:    "username",
			Usage:   "manually specify username; otherwise, will attempt to retrieve from docker store",
			EnvVars: []string{"DOCKIT_USERNAME", "DOCKIT_GRANT_USERNAME"},
		},
		&cli.StringFlag{
			Name:    "password",
			Usage:   "manually specify password; otherwise, will attempt to retrieve from docker store",
			EnvVars: []string{"DOCKIT_PASSWORD", "DOCKIT_GRANT_PASSWORD"},
		},
	}
)

func getHomeDir() string {
	if homeDir == "" {
		homeDir = homedir.Get()
	}
	return homeDir
}

func getCredentials(c *cli.Context) (username string, password string, err error) {
	if c.String("username") == "" {
		configDir = filepath.Join(getHomeDir(), configFileDir)
		cfg, _ := config.Load(configDir)
		allcreds, _ := cfg.GetAllCredentials()

		var ok bool

		creds, ok := allcreds[c.String("registry-url")]
		if !ok {
			creds, ok = allcreds[strings.ReplaceAll(c.String("registry-url"), "http://", "https://")]
			if !ok {
				return "", "", fmt.Errorf("credentials for registry not found: %s", c.String("registry-url"))
			}
		}

		username = creds.Username
		password = creds.Password
	} else {
		username = c.String("username")
		password = c.String("password")
	}

	return username, password, err
}
