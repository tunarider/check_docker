package exit

import (
	"fmt"
	"github.com/tunarider/nagios-go-sdk/nagios"
	"github.com/urfave/cli/v2"
)

type ExitForNagios func(string) cli.ExitCoder

func exitWithOutput(msg string, act func(string) nagios.Output) cli.ExitCoder {
	o := act(msg)
	fmt.Println(o.Message)
	return cli.Exit("", o.State.Int())
}

func Ok(msg string) cli.ExitCoder {
	return exitWithOutput(msg, nagios.Ok)
}

func Warning(msg string) cli.ExitCoder {
	return exitWithOutput(msg, nagios.Warning)
}

func Critical(msg string) cli.ExitCoder {
	return exitWithOutput(msg, nagios.Critical)
}

func Unknown(msg string) cli.ExitCoder {
	return exitWithOutput(msg, nagios.Unknown)
}
