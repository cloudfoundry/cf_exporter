package filters

import (
	"errors"
	"fmt"
)

const (
	ApplicationsCollector      = "Applications"
	ApplicationEventsCollector = "ApplicationEvents"
	OrganizationsCollector     = "Organizations"
	SecurityGroupsCollector    = "SecurityGroups"
	ServiceInstancesCollector  = "ServiceInstances"
	ServicesCollector          = "Services"
	SpacesCollector            = "Spaces"
)

type CollectorsFilter struct {
	collectorsEnabled map[string]bool
}

func NewCollectorsFilter(filters []string) (*CollectorsFilter, error) {
	collectorsEnabled := make(map[string]bool)

	for _, collectorName := range filters {
		switch collectorName {
		case ApplicationsCollector:
			collectorsEnabled[ApplicationsCollector] = true
		case ApplicationEventsCollector:
			collectorsEnabled[ApplicationEventsCollector] = true
		case OrganizationsCollector:
			collectorsEnabled[OrganizationsCollector] = true
		case SecurityGroupsCollector:
			collectorsEnabled[SecurityGroupsCollector] = true
		case ServiceInstancesCollector:
			collectorsEnabled[ServiceInstancesCollector] = true
		case ServicesCollector:
			collectorsEnabled[ServicesCollector] = true
		case SpacesCollector:
			collectorsEnabled[SpacesCollector] = true
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
