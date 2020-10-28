package main

import (
	"github.com/tunarider/check_docker/internal/cmd"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := cli.App{
		Name:  "check_docker",
		Usage: "Check docker status",
		Commands: []*cli.Command{
			{
				Name:    "node",
				Aliases: []string{"c"},
				Action:  cmd.Node,
			},
			{
				Name:    "service",
				Aliases: []string{"s"},
				Action:  cmd.Service,
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
