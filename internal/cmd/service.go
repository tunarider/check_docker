package cmd

import (
	"context"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/tunarider/check_docker/internal/check"
	"github.com/tunarider/check_docker/internal/exit"
	"github.com/tunarider/check_docker/internal/renderer"
	"github.com/tunarider/check_docker/internal/util"
	"github.com/tunarider/nagios-go-sdk/nagios"
	"github.com/urfave/cli/v2"
)

func notOkServiceMessage(services interface{}, performances []nagios.Performance) string {
	s := services.([]swarm.Service)
	return nagios.MessageWithPerformance(
		strings.Join(renderer.OutputFromPerformances(filterPerformances(s, performances)), ", "),
		performances,
	)
}

func filterPerformances(services []swarm.Service, performances []nagios.Performance) []nagios.Performance {
	var ps []nagios.Performance
	for _, p := range performances {
		if inService(services, p) {
			ps = append(ps, p)
		}
	}
	return ps
}

func inService(services []swarm.Service, performance nagios.Performance) bool {
	for _, s := range services {
		if s.Spec.Name == performance.Label {
			return true
		}
	}
	return false
}

type serviceResultRenderer func([]swarm.Service, []nagios.Performance) cli.ExitCoder

func serviceRenderer(exitFunc exit.ExitForNagios, msgResolver renderer.MessageResolver) serviceResultRenderer {
	return func(services []swarm.Service, performances []nagios.Performance) cli.ExitCoder {
		return exitFunc(msgResolver(services, performances))
	}
}

func getServiceRendererFunc(state nagios.State) (exit.ExitForNagios, renderer.MessageResolver) {
	switch state {
	case nagios.StateOk:
		return exit.Ok, renderer.NoProblemMessage
	case nagios.StateWarning:
		return exit.Warning, notOkServiceMessage
	case nagios.StateCritical:
		return exit.Critical, notOkServiceMessage
	default:
		return exit.Unknown, notOkServiceMessage
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
	getDesiredTasks := check.DesiredTaskGetter(ctx, dc)

	nodes, err := dc.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		return exit.Unknown("Failed to receive Docker node list")
	}
	filterExpectedNode := util.ConstraintFilter(nodes)

	state, badServices, performances := check.ServiceStatus(services, getDesiredTasks, filterExpectedNode)
	rdr := serviceRenderer(getServiceRendererFunc(state))
	return rdr(badServices, performances)
}
