package rbac

import (
	"github.com/urfave/cli/v2"

	"github.com/ekristen/dockit/pkg/commands/global"
	"github.com/ekristen/dockit/pkg/common"
)

func init() {
	flags := []cli.Flag{
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
			Usage:   "manually specify username, otherwise will attempt to retrieve from docker store",
			EnvVars: []string{"DOCKIT_USERNAME", "DOCKIT_GRANT_USERNAME"},
		},
		&cli.StringFlag{
			Name:    "password",
			Usage:   "manually specify password, otherwise will attempt to retrieve from docker store",
			EnvVars: []string{"DOCKIT_PASSWORD", "DOCKIT_GRANT_PASSWORD"},
		},
	}

	// grant user repository name action

	cliCmd := &cli.Command{
		Name:        "rbac",
		Usage:       "provides the ability to perform various RBAC related actions",
		Flags:       append(flags, global.Flags()...),
		Before:      global.Before,
		Subcommands: common.GetSubcommands("rbac"),
	}

	common.RegisterCommand(cliCmd)
}
