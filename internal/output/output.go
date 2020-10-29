package output

import (
	"fmt"
	"github.com/tunarider/check_docker/pkg/nagios"
	"github.com/urfave/cli/v2"
	"strings"
)

func Ok(msg string) cli.ExitCoder {
	return cli.Exit(fmt.Sprintf("OK: %s", msg), nagios.StateToInt(nagios.StateOk))
}

func Warning(msg string) cli.ExitCoder {
	return cli.Exit(fmt.Sprintf("Waring: %s", msg), nagios.StateToInt(nagios.StateWarning))
}

func Critical(msg string) cli.ExitCoder {
	return cli.Exit(fmt.Sprintf("Critical: %s", msg), nagios.StateToInt(nagios.StateCritical))
}

func Unknown(msg string) cli.ExitCoder {
	return cli.Exit(fmt.Sprintf("Unknown: %s", msg), nagios.StateToInt(nagios.StateUnknown))
}

func MakeOutput(msg string, performances []string) string {
	return fmt.Sprintf("%s | %s ", msg, strings.Join(performances, " "))
}
