package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/kidonchu/gitcli/command"
)

// GlobalFlags is to be used as global flag
var GlobalFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "b,branch",
		Value: "",
		Usage: "new `BRANCH` name",
	},
	cli.StringFlag{
		Name:  "s,source",
		Value: "",
		Usage: "source `BRANCH` name",
	},
	cli.StringFlag{
		Name:  "p,pattern",
		Value: "",
		Usage: "Perl/Python compatible regex `PATTERN` for finding branches and databases",
	},
}

// Commands specifies available commands
var Commands = []cli.Command{
	{
		Name:  "story",
		Usage: "story-related tasks",
		Subcommands: []cli.Command{
			// {
			// 	Name:    "new",
			// 	Aliases: []string{"n"},
			// 	Usage:   "create a new story",
			// 	Action:  command.CmdNewStory,
			// 	Flags:   GlobalFlags,
			// },
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Delete a story and its databases",
				Action:  command.CmdDeleteStory,
				Flags:   GlobalFlags,
			},
		},
	},
}

// CommandNotFound prints out the error message if command not found
func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
