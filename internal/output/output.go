package output

import (
	"fmt"
	"github.com/tunarider/check_docker/pkg/nagios"
	"github.com/urfave/cli/v2"
)

func Ok(msg string) cli.ExitCoder {
	return cli.Exit(fmt.Sprintf("OK: %s", msg), nagios.StateOk.Int())
}

func Warning(msg string) cli.ExitCoder {
	return cli.Exit(fmt.Sprintf("Waring: %s", msg), nagios.StateWarning.Int())
}

func Critical(msg string) cli.ExitCoder {
	return cli.Exit(fmt.Sprintf("Critical: %s", msg), nagios.StateCritical.Int())
}

func Unknown(msg string) cli.ExitCoder {
	return cli.Exit(fmt.Sprintf("Unknown: %s", msg), nagios.StateUnknown.Int())
}
