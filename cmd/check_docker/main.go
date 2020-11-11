package main

import (
	"github.com/tunarider/check_docker/internal/cmd"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	serviceFlags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "exclude",
			Aliases: []string{"e"},
		},
	}
	app := cli.App{
		Name:    "check_docker",
		Usage:   "Check docker status",
		Version: "v0.1.0",
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
				Flags:   serviceFlags,
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
