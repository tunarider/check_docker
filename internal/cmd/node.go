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
	"strings"
)

func listNodeNames(nodes []swarm.Node) []string {
	var nodeNames []string
	for _, node := range nodes {
		nodeNames = append(
			nodeNames,
			fmt.Sprintf("%s(%s|%s)", node.Spec.Name, node.Status, node.Spec.Availability),
		)
	}
	return nodeNames
}

type nodeMessageResolver func([]swarm.Node, []nagios.Performance) string

func okNodeMessage(_ []swarm.Node, performances []nagios.Performance) string {
	return nagios.MessageWithPerformance(
		"No problem",
		performances,
	)
}

func notOkNodeMessage(nodes []swarm.Node, performances []nagios.Performance) string {
	return nagios.MessageWithPerformance(
		strings.Join(listNodeNames(nodes), ", "),
		performances,
	)
}

type nodeResultRenderer func([]swarm.Node, []nagios.Performance) cli.ExitCoder

func nodeRenderer(exitFunc exit.ExitForNagios, msgResolver nodeMessageResolver) nodeResultRenderer {
	return func(nodes []swarm.Node, performances []nagios.Performance) cli.ExitCoder {
		return exitFunc(msgResolver(nodes, performances))
	}
}

func getNodeRenderer(state nagios.State) nodeResultRenderer {
	switch state {
	case nagios.StateOk:
		return nodeRenderer(exit.Ok, okNodeMessage)
	case nagios.StateWarning:
		return nodeRenderer(exit.Warning, notOkNodeMessage)
	case nagios.StateCritical:
		return nodeRenderer(exit.Critical, notOkNodeMessage)
	default:
		return nodeRenderer(exit.Unknown, notOkNodeMessage)
	}
}

func Node(c *cli.Context) error {
	ctx := context.Background()
	dc, err := client.NewEnvClient()
	if err != nil {
		return exit.Unknown("Failed to connect to Docker")
	}
	nodes, err := dc.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		return exit.Unknown("Failed to receive Docker node list")
	}
	state, badNodes, performances := check.CheckNodes(nodes)
	rdr := getNodeRenderer(state)
	return rdr(badNodes, performances)
}
