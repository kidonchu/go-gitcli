package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Author = "kidonchu"
	app.Email = "kidonchu@gmail.com"
	app.Usage = "Extend git cli for story"

	app.Flags = GlobalFlags
	app.Commands = Commands
	app.CommandNotFound = CommandNotFound

	app.Run(os.Args)
}
