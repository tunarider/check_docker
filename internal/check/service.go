package check

import (
	"github.com/docker/docker/api/types/swarm"
	"github.com/tunarider/check_docker/internal/util"
	"github.com/tunarider/nagios-go-sdk/nagios"
)

func makeServiceResult(service swarm.Service, desire int, actual int) (nagios.State, nagios.Performance) {
	p := nagios.Performance{
		Label:    service.Spec.Name,
		Value:    0,
		Warning:  0,
		Critical: 0,
		Min:      0,
		Max:      desire,
	}
	p.Value = actual
	if actual < desire {
		return nagios.StateCritical, p
	} else if actual > desire {
		return nagios.StateWarning, p
	} else {
		return nagios.StateOk, p
	}
}

type servicesChecker func([]swarm.Service) (nagios.State, []swarm.Service, []nagios.Performance)

func MakeServicesChecker(getTasks util.TaskGetter, filterExpectedNodes util.ExpectedNodeFilter) servicesChecker {
	return func(services []swarm.Service) (nagios.State, []swarm.Service, []nagios.Performance) {
		var state = nagios.StateOk
		var badServices []swarm.Service
		var performances []nagios.Performance
		for _, service := range services {
			var desire, actual int
			var s nagios.State
			var p nagios.Performance

			if service.Spec.Mode.Global == nil {
				desire = int(*service.Spec.Mode.Replicated.Replicas)

			} else {
				desire = len(filterExpectedNodes(service))
			}

			tasks, err := getTasks(service)
			if err != nil {
				s, p = makeServiceResult(service, desire, 0)
			} else {
				tasks = util.FilterRunningTasks(tasks)
				actual = len(tasks)
				s, p = makeServiceResult(service, desire, actual)
			}

			if s != nagios.StateOk {
				badServices = append(badServices, service)
			}
			performances = append(performances, p)
			state = nagios.ResolveState(state, s)
		}
		return state, badServices, performances
	}
}
