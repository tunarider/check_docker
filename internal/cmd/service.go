package cmd

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/tunarider/check_docker/internal/output"
	"github.com/tunarider/check_docker/pkg/nagios"
	"github.com/urfave/cli/v2"
	"strings"
	"regexp"
)

func GetServiceFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name: "filter",
			Aliases: []string{"f"},
		},
		&cli.StringSliceFlag{
			Name: "exclude",
			Aliases: []string{"e"},
		},
	}
}

func Service(c *cli.Context) error {
	var services []swarm.Service
	cf := c.String("filter")
	ex := c.StringSlice("exclude")
	ctx := context.Background()
	dc, err := client.NewEnvClient()
	ctx = context.WithValue(ctx, "dc", dc)
	if err != nil {
		return output.Unknown("Failed to connect to Docker")
	}
	f := filters.NewArgs()
	f.Add("name", cf)
	s, err := dc.ServiceList(ctx, types.ServiceListOptions{Filters: f})
	if err != nil {
		return output.Unknown("Failed to receive Docker service list")
	}
	for _, service := range s {
		if !isExclude(service, ex) {
			services = append(services, service)
		}
	} 
	state, o := checkServiceStatus(ctx, services)
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

func isExclude(service swarm.Service, excludes []string) bool {
	for _, exclude := range excludes {
		match, _ := regexp.MatchString(exclude, service.Spec.Name)
		if match {
			return true
		}
	}
	return false
}

func checkServiceStatus(ctx context.Context, services []swarm.Service) (nagios.State, string) {
	var messages []string
	var performances []string
	state := nagios.StateOk
	i := ctx.Value("dc")
	dc, ok := i.(*client.Client)
	if !ok {
		return nagios.StateUnknown, "Context error"
	}
	for _, service := range services {
		if service.Spec.Mode.Global != nil {
			continue
		}
		f := filters.NewArgs()
		f.Add("service", service.Spec.Name)
		f.Add("desired-state", "running")
		tasks, err := dc.TaskList(ctx, types.TaskListOptions{Filters: f})
		if err != nil {
			return nagios.StateUnknown, "Failed to receive Docker task list"
		}
		runningTasks := filterTask(tasks)
		desired := int(*service.Spec.Mode.Replicated.Replicas)
		actual := len(runningTasks)
		if desired > actual {
			messages = append(messages, makeServiceMessage(service, actual, desired))
			state = nagios.ResolveState(state, nagios.StateCritical)
		} else if desired < actual {
			messages = append(messages, makeServiceMessage(service, actual, desired))
			state = nagios.ResolveState(state, nagios.StateWarning)
		}
		performances = append(performances, makeServicePerformance(service, actual, desired))
	}
	o := output.MakeOutput(strings.Join(messages, ", "), performances)
	return state, o
}

func filterTask(tasks []swarm.Task) (filtered []swarm.Task) {
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

func makeServiceMessage(service swarm.Service, actual int, desired int) string {
	return fmt.Sprintf("%s(%d/%d)", service.Spec.Name, actual, desired)
}

func makeServicePerformance(service swarm.Service, actual int, desired int) string {
	return fmt.Sprintf("%s=%d;;;%d;%d", service.Spec.Name, actual, 0, desired)
}
