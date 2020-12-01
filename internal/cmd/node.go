package cmd

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/tunarider/check_docker/internal/check"
	"github.com/tunarider/check_docker/internal/exit"
	"github.com/tunarider/check_docker/internal/renderer"
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

func notOkNodeMessage(nodes interface{}, performances []nagios.Performance) string {
	n := nodes.([]swarm.Node)
	return nagios.MessageWithPerformance(
		strings.Join(listNodeNames(n), ", "),
		performances,
	)
}

type nodeResultRenderer func([]swarm.Node, []nagios.Performance) cli.ExitCoder

func nodeRenderer(exitFunc exit.ExitForNagios, msgResolver renderer.MessageResolver) nodeResultRenderer {
	return func(nodes []swarm.Node, performances []nagios.Performance) cli.ExitCoder {
		return exitFunc(msgResolver(nodes, performances))
	}
}

func getNodeRendererFunc(state nagios.State) (exit.ExitForNagios, renderer.MessageResolver) {
	switch state {
	case nagios.StateOk:
		return exit.Ok, renderer.NoProblemMessage
	case nagios.StateWarning:
		return exit.Warning, notOkNodeMessage
	case nagios.StateCritical:
		return exit.Critical, notOkNodeMessage
	default:
		return exit.Unknown, notOkNodeMessage
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
	state, badNodes, performances := check.Nodes(nodes)
	rdr := nodeRenderer(getNodeRendererFunc(state))
	return rdr(badNodes, performances)
}
