package pki

import (
	"crypto"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/ekristen/dockit/pkg/common"
	"github.com/ekristen/dockit/pkg/utils"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/dockit/pkg/commands/global"
)

type pkiCommand struct{}

func (s *pkiCommand) Execute(c *cli.Context) (err error) {
	var pem []byte
	var pub crypto.PublicKey
	var key crypto.PrivateKey

	if c.String("key-type") == "ec" {
		key1, pem1, err := utils.GenerateECKey(c.Int("key-size"))
		if err != nil {
			return err
		}
		pem = pem1
		pub = &key1.PublicKey
		key = key1
	} else if c.String("key-type") == "rsa" {
		key1, pem1, err := utils.GenerateRSAKey(c.Int("key-size"))
		if err != nil {
			return err
		}
		pem = pem1
		pub = &key1.PublicKey
		key = key1
	}

	node, err := snowflake.NewNode(1)
	if err != nil {
		return err
	}

	_, certPem, err := utils.GenerateCertificate(node.Generate().Int64(), pub, key, c.Int("years"), c.Int("months"), 0)
	if err != nil {
		return err
	}

	fmt.Println(string(pem))

	fmt.Println(string(certPem))

	return nil
}

func init() {
	cmd := pkiCommand{}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "key-type",
			Value: "rsa",
		},
		&cli.IntFlag{
			Name:  "years",
			Usage: "How long the certificate is good for",
			Value: 1,
		},
		&cli.IntFlag{
			Name:  "months",
			Usage: "The API Key used to authenticate to the SMM API",
			Value: 0,
		},
	}

	cliCmd := &cli.Command{
		Name:   "pki-generate",
		Usage:  "generates an ecdsa private key and certificate",
		Action: cmd.Execute,
		Flags:  append(flags, global.Flags()...),
		Before: global.Before,
	}

	common.RegisterCommand(cliCmd)
}
