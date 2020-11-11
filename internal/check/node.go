package check

import (
	"github.com/docker/docker/api/types/swarm"
	"github.com/tunarider/nagios-go-sdk/nagios"
)

func checkNode(node swarm.Node) (nagios.State, nagios.Performance) {
	p := nagios.Performance{
		Label:    node.Description.Hostname,
		Warning:  0,
		Critical: 0,
		Min:      0,
		Max:      2,
	}
	if node.Status.State != swarm.NodeStateReady {
		p.Value = 0
		return nagios.StateCritical, p
	}
	if node.Spec.Availability != swarm.NodeAvailabilityActive {
		p.Value = 1
		return nagios.StateWarning, p
	}
	p.Value = 2
	return nagios.StateOk, p
}

func CheckNodes(nodes []swarm.Node) (nagios.State, []swarm.Node, []nagios.Performance) {
	var state = nagios.StateOk
	var badNodes []swarm.Node
	var performances []nagios.Performance
	for _, n := range nodes {
		s, p := checkNode(n)
		if s != nagios.StateOk {
			badNodes = append(badNodes, n)
		}
		performances = append(performances, p)
		state = nagios.ResolveState(state, s)
	}
	return state, badNodes, performances
}
