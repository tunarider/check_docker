package check

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/tunarider/nagios-go-sdk/nagios"
	"math"
	"net"
)

func checkNetwork(network types.NetworkResource, inspector networkInspector, warning float64, critical float64) (nagios.State, nagios.Performance) {
	p := nagios.Performance{
		Label:    network.Name,
		Value:    0,
		Warning:  0,
		Critical: 0,
		Min:      0,
		Max:      0,
	}
	n, err := inspector(network)
	if err != nil {
		return nagios.StateUnknown, p
	}
	for _, s := range n.Services {
		p.Value += len(s.Tasks) + 1
	}
	_, ipnet, err := net.ParseCIDR(n.IPAM.Config[0].Subnet)
	if err != nil {
		return nagios.StateUnknown, p
	}
	ones, bits := ipnet.Mask.Size()
	p.Max = int(math.Pow(2, float64(bits-ones)))
	p.Warning = int(float64(p.Max) * warning)
	p.Critical = int(float64(p.Max) * critical)
	if p.Value >= p.Critical {
		return nagios.StateCritical, p
	} else if p.Value >= p.Warning {
		return nagios.StateWarning, p
	} else {
		return nagios.StateOk, p
	}
}

type networkInspector func(types.NetworkResource) (types.NetworkResource, error)

func NetworkInspector(ctx context.Context, dc *client.Client) networkInspector {
	return func(network types.NetworkResource) (types.NetworkResource, error) {
		return dc.NetworkInspect(ctx, network.ID, types.NetworkInspectOptions{Verbose: true})
	}
}

func Networks(networks []types.NetworkResource, inspector networkInspector, warning float64, critical float64) (nagios.State, []types.NetworkResource, []nagios.Performance) {
	var state = nagios.StateOk
	var badNetworks []types.NetworkResource
	var performances []nagios.Performance
	for _, n := range networks {
		s, p := checkNetwork(n, inspector, warning, critical)
		if s != nagios.StateOk {
			badNetworks = append(badNetworks, n)
		}
		performances = append(performances, p)
		state = nagios.ResolveState(state, s)
	}
	return state, badNetworks, performances
}
