package check

import (
	"math"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/tunarider/check_docker/internal/util"
	"github.com/tunarider/nagios-go-sdk/nagios"
)

func checkNetwork(network types.NetworkResource, alpha int, warning float64, critical float64) (nagios.State, nagios.Performance) {
	p := nagios.Performance{
		Label:    network.Name,
		Value:    0,
		Warning:  0,
		Critical: 0,
		Min:      0,
		Max:      0,
	}

	p.Value = alpha

	for _, s := range network.Services {
		p.Value += len(s.Tasks) + 1
	}

	_, ipnet, err := net.ParseCIDR(network.IPAM.Config[0].Subnet)
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

type networksChecker func(networks []types.NetworkResource) (nagios.State, []types.NetworkResource, []nagios.Performance)

func MakeNetworksChecker(serviceNetworkFilter util.ServiceNetworkFilter, warning float64, critical float64) networksChecker {
	return func(networks []types.NetworkResource) (nagios.State, []types.NetworkResource, []nagios.Performance) {
		var state = nagios.StateOk
		var badNetworks []types.NetworkResource
		var performances []nagios.Performance
		for _, n := range networks {
			a := 1 + len(n.Peers) + len(serviceNetworkFilter(n))
			s, p := checkNetwork(n, a, warning, critical)
			if s != nagios.StateOk {
				badNetworks = append(badNetworks, n)
			}
			performances = append(performances, p)
			state = nagios.ResolveState(state, s)
		}
		return state, badNetworks, performances
	}
}
