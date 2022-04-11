package common

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var commands []*cli.Command

var subcommands map[string][]*cli.Command = make(map[string][]*cli.Command, 0)

// Commander --
type Commander interface {
	Execute(c *cli.Context)
}

// RegisterCommand --
func RegisterCommand(command *cli.Command) {
	logrus.Debugln("Registering", command.Name, "command...")
	commands = append(commands, command)
}

func RegisterSubcommand(group string, command *cli.Command) {
	logrus.Debugln("Registering", command.Name, "command...")
	subcommands[group] = append(subcommands[group], command)
}

// GetCommands --
func GetCommands() []*cli.Command {
	return commands
}

func GetSubcommands(group string) []*cli.Command {
	return subcommands[group]
}
