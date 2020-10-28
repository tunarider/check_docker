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
	var hosts []string
	var performances []string
	for _, node := range nodes {
		nodeState := checkNodeState(node)
		nodeAvailavility := checkNodeAvailability(node)
		s := nagios.ResolveState(nodeState, nodeAvailavility)
		if s != nagios.StateOk {
			hosts = append(
				hosts,
				fmt.Sprintf(
					"%s(%s|%s)",
					node.Description.Hostname,
					node.Status.State,
					node.Spec.Availability,
				),
			)

		} else {
			performances = append(
				performances,
				fmt.Sprintf("%s=%d;%d;%d;%d;%d", node.Description.Hostname, 1, 0, 0, 0, 1),
			)
		}
		state = nagios.ResolveState(state, s)
	}
	o := fmt.Sprintf(
		"%s | %s ",
		strings.Join(hosts, ", "),
		strings.Join(performances, " "),
	)
	return state, o
}
