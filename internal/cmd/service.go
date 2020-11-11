package cmd

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/tunarider/check_docker/internal/check"
	"github.com/tunarider/check_docker/internal/exit"
	"github.com/tunarider/nagios-go-sdk/nagios"
	"github.com/urfave/cli/v2"
	"regexp"
	"strings"
)

func listServerNames(performances []nagios.Performance) []string {
	var serviceNames []string
	for _, p := range performances {
		serviceNames = append(
			serviceNames,
			fmt.Sprintf("%s(%d/%d)", p.Label, p.Value, p.Max),
		)
	}
	return serviceNames
}

type serviceMessageResolver func([]nagios.Performance) string

func okServiceMessage(performances []nagios.Performance) string {
	return nagios.MessageWithPerformance(
		"No problem",
		performances,
	)
}

func notOkServiceMessage(performances []nagios.Performance) string {
	return nagios.MessageWithPerformance(
		strings.Join(listServerNames(performances), ", "),
		performances,
	)
}

type serviceResultRenderer func([]swarm.Service, []nagios.Performance) cli.ExitCoder

func serviceRenderer(exitFunc exit.ExitForNagios, msgResolver serviceMessageResolver) serviceResultRenderer {
	return func(services []swarm.Service, performances []nagios.Performance) cli.ExitCoder {
		return exitFunc(msgResolver(performances))
	}
}

func getServiceRenderer(state nagios.State) serviceResultRenderer {
	switch state {
	case nagios.StateOk:
		return serviceRenderer(exit.Ok, okServiceMessage)
	case nagios.StateWarning:
		return serviceRenderer(exit.Warning, notOkServiceMessage)
	case nagios.StateCritical:
		return serviceRenderer(exit.Critical, notOkServiceMessage)
	default:
		return serviceRenderer(exit.Unknown, notOkServiceMessage)
	}
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

func filterService(services []swarm.Service, excludes []string) []swarm.Service {
	var filtered []swarm.Service
	for _, service := range services {
		if !isExclude(service, excludes) {
			filtered = append(filtered, service)
		}
	}
	return filtered
}

func Service(c *cli.Context) error {
	ctx := context.Background()
	dc, err := client.NewEnvClient()
	if err != nil {
		return exit.Unknown("Failed to connect to Docker")
	}
	services, err := dc.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return exit.Unknown("Failed to receive Docker service list")
	}
	services = filterService(services, c.StringSlice("exclude"))
	getRunngingTasks := check.RunningTaskGetter(ctx, dc)
	state, badServices, performances := check.CheckServiceStatus(services, getRunngingTasks)
	rdr := getServiceRenderer(state)
	return rdr(badServices, performances)
}

//func checkServiceStatus(ctx context.Context, services []swarm.Service) (nagios.State, string) {
//	var messages []string
//	var performances []string
//	state := nagios.StateOk
//	i := ctx.Value("dc")
//	dc, ok := i.(*client.Client)
//	if !ok {
//		return nagios.StateUnknown, "Context error"
//	}
//	for _, service := range services {
//		if service.Spec.Mode.Global != nil {
//			continue
//		}
//		f := filters.NewArgs()
//		f.Add("service", service.Spec.Name)
//		f.Add("desired-state", "running")
//		tasks, err := dc.TaskList(ctx, types.TaskListOptions{Filters: f})
//		if err != nil {
//			return nagios.StateUnknown, "Failed to receive Docker task list"
//		}
//		runningTasks := filterTask(tasks)
//		desired := int(*service.Spec.Mode.Replicated.Replicas)
//		actual := len(runningTasks)
//		if desired > actual {
//			messages = append(messages, makeServiceMessage(service, actual, desired))
//			state = nagios.ResolveState(state, nagios.StateCritical)
//		} else if desired < actual {
//			messages = append(messages, makeServiceMessage(service, actual, desired))
//			state = nagios.ResolveState(state, nagios.StateWarning)
//		}
//		performances = append(performances, makeServicePerformance(service, actual, desired))
//	}
//	o := exit.MakeOutput(strings.Join(messages, ", "), performances)
//	return state, o
//}
//
//func filterTask(tasks []swarm.Task) (filtered []swarm.Task) {
//	for _, task := range tasks {
//		if task.Status.State == swarm.TaskStateRunning {
//			filtered = append(filtered, task)
//		}
//	}
//	return filtered
//}
//
//func makeServiceMessage(service swarm.Service, actual int, desired int) string {
//	return fmt.Sprintf("%s(%d/%d)", service.Spec.Name, actual, desired)
//}
//
//func makeServicePerformance(service swarm.Service, actual int, desired int) string {
//	return fmt.Sprintf("%s=%d;;;%d;%d", service.Spec.Name, actual, 0, desired)
//}
