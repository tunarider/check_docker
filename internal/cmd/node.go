package cmd

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/tunarider/check_docker/internal/output"
	"github.com/tunarider/check_docker/pkg/nagios"
	"github.com/urfave/cli/v2"
	"strings"
)

func Node(c *cli.Context) error {
	ctx := context.Background()
	dc, err := client.NewEnvClient()
	if err != nil {
		return output.Unknown("Failed to connect to Docker")
	}
	nodes, err := dc.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		return output.Unknown("Failed to receive Docker node list")
	}
	state, o := checkNagiosState(nodes)
	switch state {
	case nagios.StateOk:
		return output.Ok(fmt.Sprintf("No problem%s", o))
	case nagios.StateWarning:
		return output.Warning(o)
	case nagios.StateCritical:
		return output.Critical(o)
	}
	return nil
}

func checkNodeState(node swarm.Node) nagios.State {
	if node.Status.State == swarm.NodeStateReady {
		return nagios.StateOk
	} else {
		return nagios.StateCritical
	}
}

func checkNodeAvailability(node swarm.Node) nagios.State {
	if node.Spec.Availability == swarm.NodeAvailabilityActive {
		return nagios.StateOk
	} else {
		return nagios.StateWarning
	}
}

func checkNagiosState(nodes []swarm.Node) (nagios.State, string) {
	state := nagios.StateOk
	var messages []string
	var performances []string
	for _, node := range nodes {
		nodeState := checkNodeState(node)
		nodeAvailavility := checkNodeAvailability(node)
		s := nagios.ResolveState(nodeState, nodeAvailavility)
		if s != nagios.StateOk {
			messages = append(messages, makeNodeMessage(node))
			performances = append(performances, makeNodePerformance(node, 0))
		} else {
			performances = append(performances, makeNodePerformance(node, 1))
		}
		state = nagios.ResolveState(state, s)
	}
	o := output.MakeOutput(strings.Join(messages, ", "), performances)
	return state, o
}

func makeNodeMessage(node swarm.Node) string {
	return fmt.Sprintf(
		"%s(%s|%s)",
		node.Description.Hostname,
		node.Status.State,
		node.Spec.Availability,
	)
}

func makeNodePerformance(node swarm.Node, status int) string {
	return fmt.Sprintf("%s=%d;;;%d;%d", node.Description.Hostname, status, 0, 1)
}
