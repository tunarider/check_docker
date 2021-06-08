package util

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/tunarider/check_docker/internal/exit"
)

func GetEmptyService(ctx context.Context, dc *client.Client) (emptyServices []swarm.Service, err error) {
	getTasks := MakeDesiredTaskGetter(ctx, dc)
	services, err := dc.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return emptyServices, errors.New("Failed to receive Docker service list")
	}

	for _, service := range services {
		tasks, err := getTasks(service)
		if err != nil {
			return emptyServices, exit.Unknown("Failed to receive task list")
		}
		if len(tasks) == 0 {
			emptyServices = append(emptyServices, service)
		}
	}
	return emptyServices, err
}

type ServiceNetworkFilter func(network types.NetworkResource) []swarm.Service

func MakeSerivceNetworkFilter(services []swarm.Service) ServiceNetworkFilter {
	return func(network types.NetworkResource) []swarm.Service {
		var filteredServices []swarm.Service
		for _, s := range services {
			for _, sn := range s.Spec.TaskTemplate.Networks {
				if sn.Target == network.ID {
					filteredServices = append(filteredServices, s)
					continue
				}
			}
		}
		return filteredServices
	}
}
