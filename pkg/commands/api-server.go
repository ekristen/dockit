package commands

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/ekristen/dockit/pkg/apiserver"
	"github.com/ekristen/dockit/pkg/commands/global"
	"github.com/ekristen/dockit/pkg/common"
	"github.com/ekristen/dockit/pkg/db"
	"github.com/ekristen/dockit/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type apiServerCommand struct{}

func (s *apiServerCommand) Execute(c *cli.Context) error {
	if c.String("pki-key-type") != "ec" && c.String("pki-key-type") != "rsa" {
		return fmt.Errorf("invalid pki key type: %s", c.String("pki-key-type"))
	}

	if c.Int("node-id") < 1 || c.Int("node-id") > 1024 {
		return fmt.Errorf("node-id must be 0-1023, or 1024 for random")
	}

	nodeId := c.Int64("node-id")
	if c.Int("node-id") == 1024 {
		rand.Seed(time.Now().UnixNano())
		nodeId = int64(rand.Intn(1023))
	}

	if !c.Bool("pki-generate") {
		if _, err := os.Stat(c.Path("pki-file")); err != nil {
			return errors.Wrap(err, "unable to find specified pki-file")
		}
	}

	ctx := signals.SetupSignalHandler(context.Background())

	log := logrus.WithField("command", "api-server")

	log.Infof("version: %s, node: %d", common.AppVersion.Summary, nodeId)

	node, err := snowflake.NewNode(nodeId)
	if err != nil {
		return err
	}

	dbCtx := context.WithValue(ctx, common.ContextKeyNode, node)

	database, err := db.New(dbCtx, c.String("sql-dialect"), c.String("sql-dsn"), &gorm.Config{
		Logger: db.NewLogger(c.String("log-level")),
	})
	if err != nil {
		return err
	}

	sql := database.Clauses(clause.OnConflict{DoNothing: true}).Create(&db.User{Username: "anonymous", Password: "anonymous", Active: true})
	if sql.Error != nil {
		return errors.Wrap(err, "unable to create anonymous user")
	}

	if c.String("root-user") != "" && c.String("root-password") != "" {
		sql := database.Clauses(clause.OnConflict{DoNothing: true}).Create(&db.User{Username: c.String("root-user"), Password: c.String("root-password"), Active: true, Admin: true})
		if sql.Error != nil {
			return errors.Wrap(err, "unable to create anonymous user")
		}
	}

	if err := initPKI(c, node, database, c.Bool("pki-generate"), c.Path("pki-file")); err != nil {
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
			Name:    "node-id",
			Usage:   "Unique ID of the Node (this should be increased for each replica) 0-1023 (1024 will select a random number between 0-1023)",
			EnvVars: []string{"DOCKIT_NODE_ID", "NODE_ID"},
			Value:   1024,
		},
		&cli.BoolFlag{
			Name:    "pki-generate",
			Usage:   "whether or not to generate PKI if false, you must specify --pki-file",
			Value:   true,
			EnvVars: []string{"DOCKIT_PKI_GENERATE", "PKI_GENERATE"},
		},
		&cli.PathFlag{
			Name:    "pki-file",
			Usage:   "file to read PKI data from",
			EnvVars: []string{"DOCKIT_PKI_FILE", "PKI_FILE"},
		},
		&cli.StringFlag{
			Name:    "pki-key-type",
			Usage:   "Algorithm to use for PKI for Registry to Dockit authentication",
			Value:   "ec",
			EnvVars: []string{"DOCKIT_PKI_KEY_TYPE", "PKI_KEY_TYPE"},
		},
		&cli.IntFlag{
			Name:    "pki-ec-key-size",
			Usage:   "Elliptic Curve Key Size",
			Value:   256,
			EnvVars: []string{"DOCKIT_PKI_EC_KEY_SIZE", "PKI_EC_KEY_SIZE"},
		},
		&cli.IntFlag{
			Name:    "pki-rsa-key-size",
			Usage:   "RSA Key Size",
			Value:   4096,
			EnvVars: []string{"DOCKIT_PKI_RSA_KEY_SIZE", "PKI_RSA_KEY_SIZE"},
		},
		&cli.IntFlag{
			Name:    "pki-cert-years",
			Usage:   "The number of years that internal PKI certs are good for.",
			Value:   2,
			EnvVars: []string{"DOCKIT_PKI_CERT_YEARS", "PKI_CERT_YEARS"},
		},
		&cli.IntFlag{
			Name:    "port",
			Usage:   "Port for the HTTP Server Port",
			EnvVars: []string{"DOCKIT_PORT", "PORT"},
			Value:   4315,
		},
		&cli.IntFlag{
			Name:    "metrics-port",
			Usage:   "Port for the metrics and debug http server to listen on",
			EnvVars: []string{"METRICS_PORT", "DOCKIT_METRICS_PORT"},
			Value:   4316,
		},
		&cli.StringFlag{
			Name:    "sql-dialect",
			Usage:   "The type of sql to use, sqlite or mysql",
			EnvVars: []string{"DOCKIT_SQL_DIALECT", "SQL_DIALECT"},
			Value:   "sqlite",
		},
		&cli.StringFlag{
			Name:    "sql-dsn",
			Usage:   "The DSN to use to connect to",
			EnvVars: []string{"DOCKIT_SQL_DSN", "SQL_DSN"},
			Value:   "file:dockit.sqlite",
		},
		&cli.StringFlag{
			Name:    "root-user",
			Usage:   "Root Username",
			EnvVars: []string{"DOCKIT_ROOT_USER", "ROOT_USER"},
		},
		&cli.StringFlag{
			Name:    "root-password",
			Usage:   "Root Password",
			EnvVars: []string{"DOCKIT_ROOT_PASSWORD", "ROOT_PASSWORD"},
		},
		&cli.BoolFlag{
			Name:    "first-user-admin",
			Usage:   "Indicates if the first user to login should be made an admin",
			EnvVars: []string{"DOCKIT_FIRST_USER_ADMIN", "FIRST_USER_ADMIN"},
			Value:   true,
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

func initPKI(c *cli.Context, node *snowflake.Node, database *gorm.DB, generate bool, file string) error {
	if !generate {
		pki, err := parsePKIFile(file)
		if err != nil {
			return err
		}

		sql := database.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"private", "x509"}),
		}).Create(&db.PKI{
			ID:        pki.Cert.SerialNumber.Int64(),
			Type:      pki.Cert.PublicKeyAlgorithm.String(),
			Private:   string(pki.KeyPEM),
			X509:      string(pki.CertPEM),
			NotBefore: &pki.Cert.NotBefore,
			ExpiresAt: &pki.Cert.NotAfter,
			Active:    true,
		})
		if sql.Error != nil {
			return err
		}

		return nil
	}

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
				Type:      "ECDSA",
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

type PKIFile struct {
	CertPEM []byte
	KeyPEM  []byte
	Cert    *x509.Certificate
}

func parsePKIFile(file string) (pki *PKIFile, err error) {
	pki = &PKIFile{}

	var block *pem.Block
	var rest []byte

	rest, err = ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	for {
		block, rest = pem.Decode(rest)

		if block == nil {
			break
		}

		logrus.Debugf("detected block type: %s", block.Type)

		switch block.Type {
		case "CERTIFICATE":
			pki.CertPEM = pem.EncodeToMemory(block)

			pki.Cert, err = x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}
		case "RSA PRIVATE KEY", "EC PRIVATE KEY":
			pki.KeyPEM = pem.EncodeToMemory(block)
		}
	}

	if len(pki.CertPEM) == 0 {
		err = fmt.Errorf("unable to find certificate")
	} else if len(pki.KeyPEM) == 0 {
		err = fmt.Errorf("unable to find key")
	}

	return pki, err
}
