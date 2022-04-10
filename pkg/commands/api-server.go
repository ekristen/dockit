package commands

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"github.com/ekristen/dockit/pkg/apiserver"
	"github.com/ekristen/dockit/pkg/commands/global"
	"github.com/ekristen/dockit/pkg/common"
	"github.com/ekristen/dockit/pkg/db"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

type apiServerCommand struct{}

func (s *apiServerCommand) Execute(c *cli.Context) error {
	ctx := signals.SetupSignalHandler(context.Background())

	log := logrus.WithField("command", "api-server")

	node, err := snowflake.NewNode(1)
	if err != nil {
		return err
	}

	dbCtx := context.WithValue(ctx, common.ContextKeyNode, node)

	database, err := db.New(dbCtx, "sqlite", "dockit.sqlite", &gorm.Config{
		Logger: db.NewLogger(c.String("log-level")),
	})
	if err != nil {
		return err
	}

	apiServer := apiserver.Register(ctx, log, database, c.Int("port"))

	if err := apiServer.Start(); err != nil {
		return err
	}

	return nil
}

func init() {
	cmd := apiServerCommand{}

	flags := []cli.Flag{
		&cli.IntFlag{
			Name:    "port",
			Usage:   "Port for the HTTP Server Port",
			EnvVars: []string{"PORT"},
			Value:   4315,
		},
		&cli.IntFlag{
			Name:    "metrics-port",
			Usage:   "Port for the metrics and debug http server to listen on",
			EnvVars: []string{"METRICS_PORT", "API_SERVER_METRICS_PORT", "ODIN_API_SERVER_METRICS_PORT"},
			Value:   4316,
		},
	}

	cliCmd := &cli.Command{
		Name:   "api-server",
		Usage:  "api-server",
		Action: cmd.Execute,
		Flags:  append(flags, global.Flags()...),
		Before: global.Before,
	}

	common.RegisterCommand(cliCmd)
}
