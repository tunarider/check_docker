package output

import (
	"fmt"
	"github.com/tunarider/check_docker/pkg/nagios"
	"github.com/urfave/cli/v2"
	"strings"
)

func Ok(msg string) cli.ExitCoder {
	fmt.Printf("OK: %s", msg)
	return cli.Exit("", nagios.StateToInt(nagios.StateOk))
}

func Warning(msg string) cli.ExitCoder {
	fmt.Printf("Warning: %s", msg)
	return cli.Exit("", nagios.StateToInt(nagios.StateWarning))
}

func Critical(msg string) cli.ExitCoder {
	fmt.Printf("Critical: %s", msg)
	return cli.Exit("", nagios.StateToInt(nagios.StateCritical))
}

func Unknown(msg string) cli.ExitCoder {
	fmt.Printf("Unknown: %s", msg)
	return cli.Exit("", nagios.StateToInt(nagios.StateUnknown))
}

func MakeOutput(msg string, performances []string) string {
	return fmt.Sprintf("%s | %s ", msg, strings.Join(performances, " "))
}
