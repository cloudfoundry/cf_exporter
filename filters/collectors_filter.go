package filters

import (
	"errors"
	"fmt"
	"strings"
)

const (
	ApplicationsCollector      = "Applications"
	EventsCollector            = "Events"
	IsolationSegmentsCollector = "IsolationSegments"
	OrganizationsCollector     = "Organizations"
	RoutesCollector            = "Routes"
	SecurityGroupsCollector    = "SecurityGroups"
	ServiceBindingsCollector   = "ServiceBindings"
	ServiceInstancesCollector  = "ServiceInstances"
	ServicePlansCollector      = "ServicePlans"
	ServicesCollector          = "Services"
	SpacesCollector            = "Spaces"
	StacksCollector            = "Stacks"
)

type CollectorsFilter struct {
	collectorsEnabled map[string]bool
	CFAPIv3Enabled    bool
}

func NewCollectorsFilter(filters []string, cfAPIv3Enabled bool) (*CollectorsFilter, error) {
	collectorsEnabled := make(map[string]bool)

	for _, collectorName := range filters {
		switch strings.Trim(collectorName, " ") {
		case ApplicationsCollector:
			collectorsEnabled[ApplicationsCollector] = true
		case EventsCollector:
			collectorsEnabled[EventsCollector] = true
		case IsolationSegmentsCollector:
			if !cfAPIv3Enabled {
				return &CollectorsFilter{}, errors.New("IsolationSegments Collector filter need CF API V3 enabled")
			}
			collectorsEnabled[IsolationSegmentsCollector] = true
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
		case ServicePlansCollector:
			collectorsEnabled[ServicePlansCollector] = true
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

	return &CollectorsFilter{collectorsEnabled: collectorsEnabled, CFAPIv3Enabled: cfAPIv3Enabled}, nil
}

func (f *CollectorsFilter) Enabled(collectorName string) bool {
	if len(f.collectorsEnabled) == 0 {
		switch strings.Trim(collectorName, " ") {
		case IsolationSegmentsCollector:
			if f.CFAPIv3Enabled {
				return true
			}
			return false
		case EventsCollector:
			return false
		default:
			return true
		}
	}

	if f.collectorsEnabled[collectorName] {
		return true
	}

	return false
}
