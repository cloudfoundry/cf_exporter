package filters

import (
	"errors"
	"fmt"
)

const (
	ApplicationsCollector  = "Applications"
	OrganizationsCollector = "Organizations"
	ServicesCollector      = "Services"
	SpacesCollector        = "Spaces"
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
		case OrganizationsCollector:
			collectorsEnabled[OrganizationsCollector] = true
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
