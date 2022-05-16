package initcontainer

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ekristen/dockit/pkg/common"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/dockit/pkg/commands/global"
)

type initContainerCommand struct{}

func (s *initContainerCommand) Execute(c *cli.Context) (err error) {
	if c.Args().Len() != 1 {
		return fmt.Errorf("please specify path to write cert bundle to")
	}

	ctx := signals.SetupSignalHandler(context.Background())

	url := fmt.Sprintf("%s/certs/pem", c.String("base-url"))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	logrus.WithField("headers", resp.Header).Debug("Response Headers")
	logrus.WithField("status", resp.StatusCode).Debug("Response Status Code")

	body, _ := ioutil.ReadAll(resp.Body)

	if err := ioutil.WriteFile(c.Args().First(), body, 0644); err != nil {
		return err
	}

	return nil
}

func init() {
	cmd := initContainerCommand{}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "base-url",
			Value:   "http://localhost:4315/v2",
			EnvVars: []string{"DOCKIT_BASE_URL", "DOCKIT_INITCONTAINER_BASE_URL"},
		},
	}

	cliCmd := &cli.Command{
		Name:   "init-container",
		Usage:  "Provides init container capability to fetch and store PKI from Dockit before starting registry server",
		Action: cmd.Execute,
		Flags:  append(flags, global.Flags()...),
		Before: global.Before,
	}

	common.RegisterCommand(cliCmd)
}
