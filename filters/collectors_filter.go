package filters

import (
	"errors"
	"fmt"
	"strings"
)

const (
	ApplicationsCollector     = "Applications"
	OrganizationsCollector    = "Organizations"
	RoutesCollector           = "Routes"
	SecurityGroupsCollector   = "SecurityGroups"
	ServiceBindingsCollector  = "ServiceBindings"
	ServiceInstancesCollector = "ServiceInstances"
	ServicesCollector         = "Services"
	SpacesCollector           = "Spaces"
	StacksCollector           = "Stacks"
)

type CollectorsFilter struct {
	collectorsEnabled map[string]bool
}

func NewCollectorsFilter(filters []string) (*CollectorsFilter, error) {
	collectorsEnabled := make(map[string]bool)

	for _, collectorName := range filters {
		switch strings.Trim(collectorName, " ") {
		case ApplicationsCollector:
			collectorsEnabled[ApplicationsCollector] = true
		case OrganizationsCollector:
			collectorsEnabled[OrganizationsCollector] = true
		case RoutesCollector:
			collectorsEnabled[RoutesCollector] = true
		case SecurityGroupsCollector:
			collectorsEnabled[SecurityGroupsCollector] = true
		case ServiceBindingsCollector:
			collectorsEnabled[ServiceBindingsCollector] = true
		case ServiceInstancesCollector:
			collectorsEnabled[ServiceInstancesCollector] = true
		case ServicesCollector:
			collectorsEnabled[ServicesCollector] = true
		case SpacesCollector:
			collectorsEnabled[SpacesCollector] = true
		case StacksCollector:
			collectorsEnabled[StacksCollector] = true
		default:
			return &CollectorsFilter{}, errors.New(fmt.Sprintf("Collector filter `%s` is not supported", collectorName))
		}
	}

	return &CollectorsFilter{collectorsEnabled: collectorsEnabled}, nil
}

func (f *CollectorsFilter) Enabled(collectorName string) bool {
	if len(f.collectorsEnabled) == 0 {
		return true
	}

	if f.collectorsEnabled[collectorName] {
		return true
	}

	return false
}
