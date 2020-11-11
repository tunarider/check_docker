package cmd

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/tunarider/check_docker/internal/check"
	"github.com/tunarider/check_docker/internal/exit"
	"github.com/tunarider/nagios-go-sdk/nagios"
	"github.com/urfave/cli/v2"
	"regexp"
	"strings"
)

func listServerNames(performances []nagios.Performance) []string {
	var serviceNames []string
	for _, p := range performances {
		serviceNames = append(
			serviceNames,
			fmt.Sprintf("%s(%d/%d)", p.Label, p.Value, p.Max),
		)
	}
	return serviceNames
}

type serviceMessageResolver func([]nagios.Performance) string

func okServiceMessage(performances []nagios.Performance) string {
	return nagios.MessageWithPerformance(
		"No problem",
		performances,
	)
}

func notOkServiceMessage(performances []nagios.Performance) string {
	return nagios.MessageWithPerformance(
		strings.Join(listServerNames(performances), ", "),
		performances,
	)
}

type serviceResultRenderer func([]swarm.Service, []nagios.Performance) cli.ExitCoder

func serviceRenderer(exitFunc exit.ExitForNagios, msgResolver serviceMessageResolver) serviceResultRenderer {
	return func(services []swarm.Service, performances []nagios.Performance) cli.ExitCoder {
		return exitFunc(msgResolver(performances))
	}
}

func getServiceRenderer(state nagios.State) serviceResultRenderer {
	switch state {
	case nagios.StateOk:
		return serviceRenderer(exit.Ok, okServiceMessage)
	case nagios.StateWarning:
		return serviceRenderer(exit.Warning, notOkServiceMessage)
	case nagios.StateCritical:
		return serviceRenderer(exit.Critical, notOkServiceMessage)
	default:
		return serviceRenderer(exit.Unknown, notOkServiceMessage)
	}
}

func isExclude(service swarm.Service, excludes []string) bool {
	for _, exclude := range excludes {
		match, _ := regexp.MatchString(exclude, service.Spec.Name)
		if match {
			return true
		}
	}
	return false
}

func filterService(services []swarm.Service, excludes []string) []swarm.Service {
	var filtered []swarm.Service
	for _, service := range services {
		if !isExclude(service, excludes) {
			filtered = append(filtered, service)
		}
	}
	return filtered
}

func Service(c *cli.Context) error {
	ctx := context.Background()
	dc, err := client.NewEnvClient()
	if err != nil {
		return exit.Unknown("Failed to connect to Docker")
	}
	services, err := dc.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return exit.Unknown("Failed to receive Docker service list")
	}
	services = filterService(services, c.StringSlice("exclude"))
	getRunngingTasks := check.RunningTaskGetter(ctx, dc)
	state, badServices, performances := check.CheckServiceStatus(services, getRunngingTasks)
	rdr := getServiceRenderer(state)
	return rdr(badServices, performances)
}
