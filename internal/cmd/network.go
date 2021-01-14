package cmd

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/tunarider/check_docker/internal/check"
	"github.com/tunarider/check_docker/internal/exit"
	"github.com/tunarider/check_docker/internal/renderer"
	"github.com/tunarider/nagios-go-sdk/nagios"
	"github.com/urfave/cli/v2"
	"strings"
)

func notOkNetworkMessage(networks interface{}, performances []nagios.Performance) string {
	n := networks.([]types.NetworkResource)
	return nagios.MessageWithPerformance(
		strings.Join(renderer.OutputFromPerformances(filterNetworkPerformances(n, performances)), ","),
		performances,
	)
}

func filterNetworkPerformances(networks []types.NetworkResource, performances []nagios.Performance) []nagios.Performance {
	var ps []nagios.Performance
	for _, p := range performances {
		if inNetwork(networks, p) {
			ps = append(ps, p)
		}
	}
	return ps
}

func inNetwork(networks []types.NetworkResource, performance nagios.Performance) bool {
	for _, n := range networks {
		if n.Name == performance.Label {
			return true
		}
	}
	return false
}

type networkResultRenderer func([]types.NetworkResource, []nagios.Performance) cli.ExitCoder

func networkRenderer(exitFunc exit.ExitForNagios, msgResolver renderer.MessageResolver) networkResultRenderer {
	return func(networks []types.NetworkResource, performances []nagios.Performance) cli.ExitCoder {
		return exitFunc(msgResolver(networks, performances))
	}
}

func getNetworkRendererFunc(state nagios.State) (exit.ExitForNagios, renderer.MessageResolver) {
	switch state {
	case nagios.StateOk:
		return exit.Ok, renderer.NoProblemMessage
	case nagios.StateWarning:
		return exit.Warning, notOkNetworkMessage
	case nagios.StateCritical:
		return exit.Critical, notOkNetworkMessage
	default:
		return exit.Unknown, notOkNetworkMessage
	}
}

func Network(c *cli.Context) error {
	ctx := context.Background()
	dc, err := client.NewEnvClient()
	if err != nil {
		return exit.Unknown("Failed to connect to Docker")
	}
	f := filters.NewArgs()
	f.Add("driver", "overlay")
	networks, err := dc.NetworkList(ctx, types.NetworkListOptions{Filters: f})
	if err != nil {
		return exit.Unknown("Failed to receive Docker network list")
	}
	networkInspector := check.NetworkInspector(ctx, dc)
	state, badNetworks, performances := check.Networks(networks, networkInspector, c.Float64("warning"), c.Float64("critical"))
	rdr := networkRenderer(getNetworkRendererFunc(state))
	return rdr(badNetworks, performances)
}
