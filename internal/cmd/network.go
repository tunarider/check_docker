package cmd

import (
	"context"
	"errors"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/tunarider/check_docker/internal/check"
	"github.com/tunarider/check_docker/internal/exit"
	"github.com/tunarider/check_docker/internal/renderer"
	"github.com/tunarider/check_docker/internal/util"
	"github.com/tunarider/nagios-go-sdk/nagios"
	"github.com/urfave/cli/v2"
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

func getVerboseNetworks(ctx context.Context, dc *client.Client) ([]types.NetworkResource, error) {
	var verboseNetworks []types.NetworkResource

	f := filters.NewArgs()
	f.Add("driver", "overlay")
	networks, err := dc.NetworkList(ctx, types.NetworkListOptions{Filters: f})
	if err != nil {
		return verboseNetworks, errors.New("Failed to receive Docker network list")
	}

	for _, network := range networks {
		n, err := dc.NetworkInspect(ctx, network.ID, types.NetworkInspectOptions{Verbose: true})
		if err != nil {
			return verboseNetworks, errors.New("Failed to inspect Docker network")
		}
		verboseNetworks = append(verboseNetworks, n)
	}
	return verboseNetworks, nil
}

func Network(c *cli.Context) error {
	ctx := context.Background()
	dc, err := client.NewEnvClient()
	if err != nil {
		return exit.Unknown("Failed to connect to Docker")
	}

	networks, err := getVerboseNetworks(ctx, dc)
	if err != nil {
		return exit.Unknown(err.Error())
	}

	emptyServices, err := util.GetEmptyService(ctx, dc)
	if err != nil {
		return exit.Unknown(err.Error())
	}

	checkNetworks := check.MakeNetworksChecker(util.MakeSerivceNetworkFilter(emptyServices), c.Float64("warning"), c.Float64("critical"))

	state, badNetworks, performances := checkNetworks(networks)
	rdr := networkRenderer(getNetworkRendererFunc(state))
	return rdr(badNetworks, performances)
}
