package grant

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/dockit/pkg/commands/global"
	"github.com/ekristen/dockit/pkg/common"
)

type grantCommand struct{}

func (s *grantCommand) Execute(c *cli.Context) (err error) {
	if c.Args().Len() != 4 {
		return fmt.Errorf("incorrect usage")
	}

	url := fmt.Sprintf("%s/admin/%s/%s", c.String("base-url"), c.String("type"), strings.Join(c.Args().Slice(), "/"))
	logrus.Debug(url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	if c.String("api-token") != "" {
		req.Header.Set("x-api-token", c.String("api-token"))
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

	logrus.WithField("headers", resp.Header).Info("Response Headers")
	logrus.WithField("status", resp.StatusCode).Info("Response Status Code")

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

	return nil
}

func init() {
	cmd := grantCommand{}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "base-url",
			Value: "http://localhost:4315/v2",
		},
		&cli.StringFlag{
			Name:  "type",
			Value: "user",
		},
		&cli.BoolFlag{
			Name: "insecure",
		},
	}

	// grant user repository name action

	cliCmd := &cli.Command{
		Name:   "grant",
		Usage:  "grant access",
		Action: cmd.Execute,
		Flags:  append(flags, global.Flags()...),
		Before: global.Before,
	}

	common.RegisterCommand(cliCmd)
}
