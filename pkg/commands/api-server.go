package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/ekristen/dockit/pkg/apiserver"
	"github.com/ekristen/dockit/pkg/commands/global"
	"github.com/ekristen/dockit/pkg/common"
	"github.com/ekristen/dockit/pkg/db"
	"github.com/ekristen/dockit/pkg/utils"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

type apiServerCommand struct{}

func (s *apiServerCommand) Execute(c *cli.Context) error {
	if c.String("pki-key-type") != "ec" && c.String("pki-key-type") != "rsa" {
		return fmt.Errorf("invalid pki key type: %s", c.String("pki-key-type"))
	}

	ctx := signals.SetupSignalHandler(context.Background())

	log := logrus.WithField("command", "api-server")

	log.Infof("version: %s", common.AppVersion.Summary)

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

	if err := initPKI(c, node, database); err != nil {
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
			Name:  "node-id",
			Usage: "Unique ID of the Node (this should be increased for each replica)",
			Value: 1,
		},
		&cli.StringFlag{
			Name:  "pki-key-type",
			Value: "ec",
		},
		&cli.IntFlag{
			Name:  "pki-ec-key-size",
			Value: 256,
		},
		&cli.IntFlag{
			Name:  "pki-rsa-key-size",
			Value: 4096,
		},
		&cli.IntFlag{
			Name:  "pki-cert-years",
			Value: 2,
		},
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
		&cli.StringFlag{
			Name:  "init-cert-bundle",
			Usage: "useful in development, update cert bundle at path with latest keys",
		},
	}

	cliCmd := &cli.Command{
		Name:   "api-server",
		Usage:  "dockit api server",
		Action: cmd.Execute,
		Flags:  append(flags, global.Flags()...),
		Before: global.Before,
	}

	common.RegisterCommand(cliCmd)
}

func initPKI(c *cli.Context, node *snowflake.Node, database *gorm.DB) error {
	sql := database.Model(&db.PKI{}).Where("expires_at > ? AND active = 1", time.Now().UTC()).Find(nil)
	if sql.Error != nil {
		return sql.Error
	}
	if sql.RowsAffected == 0 {
		logrus.Info("generating pki for signing tokens")
		id := node.Generate().Int64()

		switch c.String("pki-key-type") {
		case "ec":
			key, keyPEM, err := utils.GenerateECKey(c.Int("pki-ec-key-size"))
			if err != nil {
				return err
			}
			cert, certPEM, err := utils.GenerateCertificate(id, &key.PublicKey, key, c.Int("pki-cert-years"), 0, 0)
			if err != nil {
				return err
			}

			sql := database.Create(&db.PKI{
				ID:        id,
				Type:      "ec",
				Private:   string(keyPEM),
				X509:      string(certPEM),
				Bits:      c.Int("pki-ec-key-size"),
				NotBefore: &cert.NotBefore,
				ExpiresAt: &cert.NotAfter,
				Active:    true,
			})
			if sql.Error != nil {
				return err
			}
		case "rsa":
			key, keyPEM, err := utils.GenerateRSAKey(c.Int("pki-rsa-key-size"))
			if err != nil {
				return err
			}
			cert, certPEM, err := utils.GenerateCertificate(id, key.PublicKey, key, c.Int("pki-cert-years"), 0, 0)
			if err != nil {
				return err
			}
			sql := database.Create(&db.PKI{
				ID:        id,
				Type:      "ec",
				Private:   string(keyPEM),
				X509:      string(certPEM),
				Bits:      c.Int("pki-rsa-key-size"),
				NotBefore: &cert.NotBefore,
				ExpiresAt: &cert.NotAfter,
				Active:    true,
			})
			if sql.Error != nil {
				return err
			}
		}

	}

	return nil
}
