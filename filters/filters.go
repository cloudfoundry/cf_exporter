package filters

import (
	"fmt"
	"strings"
)

const (
	Applications      = "applications"
	Buildpacks        = "buildpacks"
	Events            = "events"
	IsolationSegments = "isolationsegments"
	Organizations     = "organizations"
	Routes            = "routes"
	SecurityGroups    = "securitygroups"
	ServiceBindings   = "servicebindings"
	ServiceInstances  = "serviceinstances"
	ServicePlans      = "serviceplans"
	Services          = "services"
	Spaces            = "spaces"
	Stacks            = "stacks"
)

var (
	All = []string{
		Applications,
		Buildpacks,
		Events,
		IsolationSegments,
		Organizations,
		Routes,
		SecurityGroups,
		ServiceBindings,
		ServiceInstances,
		ServicePlans,
		Services,
		Spaces,
		Stacks,
	}
)

type Filter struct {
	activated map[string]bool
}

func NewFilter(active ...string) (*Filter, error) {
	filter := &Filter{
		activated: map[string]bool{
			Applications:      true,
			Buildpacks:        true,
			IsolationSegments: true,
			Organizations:     true,
			Routes:            true,
			SecurityGroups:    true,
			ServiceBindings:   true,
			ServiceInstances:  true,
			ServicePlans:      true,
			Services:          true,
			Spaces:            true,
			Stacks:            true,
			Events:            false,
		},
	}

	if len(active) != 0 {
		if err := filter.setActive(active); err != nil {
			return nil, err
		}
	}

	return filter, nil
}

func (f *Filter) setActive(active []string) error {
	// override default states with all disabled
	f.activated = map[string]bool{
		Applications:      false,
		Buildpacks:        false,
		IsolationSegments: false,
		Organizations:     false,
		Routes:            false,
		SecurityGroups:    false,
		ServiceBindings:   false,
		ServiceInstances:  false,
		ServicePlans:      false,
		Services:          false,
		Spaces:            false,
		Stacks:            false,
		Events:            false,
	}

	// enable only given filters
	for _, val := range active {
		name := strings.Trim(val, " ")
		name = strings.ToLower(name)
		if _, ok := f.activated[name]; !ok {
			return fmt.Errorf("Filter `%s` is not supported", val)
		}
		f.activated[name] = true
	}
	return nil
}

func (f *Filter) Enabled(name string) bool {
	status, ok := f.activated[name]
	return (ok && status)
}

func (f *Filter) Any(names ...string) bool {
	for _, n := range names {
		if f.Enabled(n) {
			return true
		}
	}
	return false
}

func (f *Filter) All(names ...string) bool {
	for _, n := range names {
		if !f.Enabled(n) {
			return false
		}
	}
	return true
}
