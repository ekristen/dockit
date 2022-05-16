package rbac

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/dockit/pkg/apiserver/response"
	"github.com/ekristen/dockit/pkg/commands/global"
	"github.com/ekristen/dockit/pkg/common"
)

type permissionCommand struct{}

func (s *permissionCommand) Execute(c *cli.Context) (err error) {
	if c.Args().Len() != 2 {
		return fmt.Errorf("incorrect usage")
	}

	username, password, err := getCredentials(c)
	if err != nil {
		return err
	}

	url1 := strings.Join(c.Args().Slice(), "|")
	url2 := strings.ReplaceAll(url1, "/", "_")
	url3 := strings.ReplaceAll(url2, "|", "/")

	url := fmt.Sprintf("%s/admin/%s", c.String("base-url"), url3)
	logrus.WithField("url", url).Debug("request url")

	method := "PUT"
	if c.Command.Name == "revoke" {
		method = "DELETE"
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	basicCreds := fmt.Sprintf("%s:%s", username, password)
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(basicCreds)))
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.Bool("insecure"),
		},
	}}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	logrus.WithField("headers", resp.Header).Debug("response headers")
	logrus.WithField("status", resp.StatusCode).Debug("response Status Code")

	res, err := response.ReadAllDecode(resp.Body)
	if err != nil {
		return err
	}

	if res.Status {
		fmt.Printf("%s successful\n", c.Command.Name)
	} else {
		fmt.Printf("%s failed\n", c.Command.Name)
	}

	return nil
}

func init() {
	cmd := permissionCommand{}

	// flags := []cli.Flag{}

	// grant user repository name action

	grantCmd := &cli.Command{
		Name:   "grant",
		Usage:  "grant (user|group):<name> (repository|namespace):<name>:(pull|push)",
		Action: cmd.Execute,
		Flags:  append(rbacFlags, global.Flags()...),
		Before: global.Before,
	}

	revokeCmd := &cli.Command{
		Name:   "revoke",
		Usage:  "revoke (user|group):<name> (repository|namespace):<name>:(pull|push)",
		Action: cmd.Execute,
		Flags:  append(rbacFlags, global.Flags()...),
		Before: global.Before,
	}

	common.RegisterSubcommand("rbac", grantCmd)
	common.RegisterSubcommand("rbac", revokeCmd)
}
