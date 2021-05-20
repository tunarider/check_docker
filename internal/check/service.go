package check

import (
	"context"
	// "fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/tunarider/nagios-go-sdk/nagios"
)

func ServiceStatus(services []swarm.Service, tg taskGetter) (nagios.State, []swarm.Service, []nagios.Performance) {
	var state = nagios.StateOk
	var badServices []swarm.Service
	var performances []nagios.Performance
	for _, service := range services {
		var s nagios.State
		var p nagios.Performance
		if service.Spec.Mode.Global == nil {
			s, p = checkReplicatedService(service, tg)
			if s != nagios.StateOk {
				badServices = append(badServices, service)
			}
			performances = append(performances, p)
			state = nagios.ResolveState(state, s)
		} else {
			s, p = checkGlobalService(service, tg)
			if s != nagios.StateOk {
				badServices = append(badServices, service)
			}
			performances = append(performances, p)
			state = nagios.ResolveState(state, s)
		}
	}
	return state, badServices, performances
}

func checkGlobalService(service swarm.Service, tg taskGetter) (nagios.State, nagios.Performance) {
	desiredTasks, err := tg(service)
	desire := len(desiredTasks)
	p := nagios.Performance{
		Label:    service.Spec.Name,
		Value:    0,
		Warning:  0,
		Critical: 0,
		Min:      0,
		Max:      desire,
	}
	tasks := filterRunningTasks(desiredTasks)
	if err != nil {
		return nagios.StateUnknown, p
	}
	actual := len(tasks)
	p.Value = actual
	if actual < desire {
		return nagios.StateCritical, p
	} else if actual > desire {
		return nagios.StateWarning, p
	} else {
		return nagios.StateOk, p
	}
}

func checkReplicatedService(service swarm.Service, tg taskGetter) (nagios.State, nagios.Performance) {
	desire := int(*service.Spec.Mode.Replicated.Replicas)
	p := nagios.Performance{
		Label:    service.Spec.Name,
		Value:    0,
		Warning:  0,
		Critical: 0,
		Min:      0,
		Max:      desire,
	}
	tasks, err := tg(service)
	if err != nil {
		return nagios.StateUnknown, p
	}
	tasks = filterRunningTasks(tasks)
	actual := len(tasks)
	p.Value = actual
	if actual < desire {
		return nagios.StateCritical, p
	} else if actual > desire {
		return nagios.StateWarning, p
	} else {
		return nagios.StateOk, p
	}
}

type taskGetter func(swarm.Service) ([]swarm.Task, error)

func DesiredTaskGetter(ctx context.Context, dc *client.Client) taskGetter {
	return func(service swarm.Service) (tasks []swarm.Task, err error) {
		f := filters.NewArgs()
		f.Add("service", service.Spec.Name)
		f.Add("desired-state", "running")
		tasks, err = dc.TaskList(ctx, types.TaskListOptions{Filters: f})
		return tasks, err
	}
}

func filterRunningTasks(tasks []swarm.Task) []swarm.Task {
	var filtered []swarm.Task
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning {
			filtered = append(filtered, task)
		}
	}
	return filtered
}
