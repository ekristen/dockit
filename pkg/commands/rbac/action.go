package rbac

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/dockit/pkg/commands/global"
	"github.com/ekristen/dockit/pkg/common"
)

type actionCommand struct{}

func (s *actionCommand) Execute(c *cli.Context) (err error) {
	switch c.Command.Name {
	case "add-member", "remove-member":
		if c.Args().Len() != 2 {
			return fmt.Errorf("usage: %s group:<name> user:<username>", c.Command.Name)
		}
	default:
		if c.Args().Len() != 1 {
			return fmt.Errorf("invalid usage, missing first argument")
		}

		parts := strings.Split(c.Args().First(), ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid rbac entity, format should be (user|group):<name>")
		}
	}

	username, password, err := getCredentials(c)
	if err != nil {
		return err
	}

	var data []byte = nil

	if c.Command.Name == "change-password" {
		p := struct {
			Password string `json:"password"`
		}{
			Password: c.Args().Get(2),
		}

		data, err = json.Marshal(p)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/admin/%s/%s", c.String("base-url"), strings.Join(c.Args().Slice(), "/"), c.Command.Name)
	logrus.WithField("url", url).Debug("request url")

	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))

	return nil
}

func init() {
	cmd := actionCommand{}

	// flags := []cli.Flag{}

	// grant user repository name action

	chpwdCmd := &cli.Command{
		Name:   "change-password",
		Usage:  "change password of a user",
		Action: cmd.Execute,
		Flags:  append(rbacFlags, global.Flags()...),
		Before: global.Before,
	}

	enableCmd := &cli.Command{
		Name:   "disable",
		Usage:  "disable a user or group",
		Action: cmd.Execute,
		Flags:  append(rbacFlags, global.Flags()...),
		Before: global.Before,
	}

	disableCmd := &cli.Command{
		Name:   "enable",
		Usage:  "enable a user or group",
		Action: cmd.Execute,
		Flags:  append(rbacFlags, global.Flags()...),
		Before: global.Before,
	}

	addCmd := &cli.Command{
		Name:   "add",
		Usage:  "add a user or group",
		Action: cmd.Execute,
		Flags:  append(rbacFlags, global.Flags()...),
		Before: global.Before,
	}

	removeCmd := &cli.Command{
		Name:   "remove",
		Usage:  "remove a user or group",
		Action: cmd.Execute,
		Flags:  append(rbacFlags, global.Flags()...),
		Before: global.Before,
	}

	addMemberCmd := &cli.Command{
		Name:   "add-member",
		Usage:  "add a user to a group",
		Action: cmd.Execute,
		Flags:  append(rbacFlags, global.Flags()...),
		Before: global.Before,
	}

	removeMemberCmd := &cli.Command{
		Name:   "remove-member",
		Usage:  "remove a user from a group",
		Action: cmd.Execute,
		Flags:  append(rbacFlags, global.Flags()...),
		Before: global.Before,
	}

	listPermissionsCmd := &cli.Command{
		Name:   "permissions",
		Usage:  "list permissions for a user or group",
		Action: cmd.Execute,
		Flags:  append(rbacFlags, global.Flags()...),
		Before: global.Before,
	}

	common.RegisterSubcommand("rbac", chpwdCmd)
	common.RegisterSubcommand("rbac", enableCmd)
	common.RegisterSubcommand("rbac", disableCmd)
	common.RegisterSubcommand("rbac", addCmd)
	common.RegisterSubcommand("rbac", removeCmd)
	common.RegisterSubcommand("rbac", addMemberCmd)
	common.RegisterSubcommand("rbac", removeMemberCmd)
	common.RegisterSubcommand("rbac", listPermissionsCmd)
}
