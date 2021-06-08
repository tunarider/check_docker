package util

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

type TaskGetter func(swarm.Service) ([]swarm.Task, error)

func MakeDesiredTaskGetter(ctx context.Context, dc *client.Client) TaskGetter {
	return func(service swarm.Service) (tasks []swarm.Task, err error) {
		f := filters.NewArgs()
		f.Add("service", service.Spec.Name)
		f.Add("desired-state", "running")
		tasks, err = dc.TaskList(ctx, types.TaskListOptions{Filters: f})
		return tasks, err
	}
}

func FilterRunningTasks(tasks []swarm.Task) []swarm.Task {
	var filtered []swarm.Task
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning {
			filtered = append(filtered, task)
		}
	}
	return filtered
}
