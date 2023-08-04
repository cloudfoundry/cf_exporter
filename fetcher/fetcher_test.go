package fetcher

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bosh-prometheus/cf_exporter/filters"
)

var _ = Describe("Fetcher", func() {
	Context("fetching jobs are planned according to filter", func() {
		var (
			fetcher  *Fetcher
			active   []string
			jobs     []string
			expected []string
		)

		JustBeforeEach(func() {
			f, err := filters.NewFilter(active...)
			Ω(err).ShouldNot(HaveOccurred())
			fetcher = NewFetcher(10, &CFConfig{}, f)
			Ω(fetcher).ShouldNot(BeNil())
			fetcher.workInit()

			close(fetcher.worker.list)
			jobs = []string{}
			for w := range fetcher.worker.list {
				jobs = append(jobs, w.name)
			}
		})

		When("default filters are set", func() {
			BeforeEach(func() {
				active = []string{}
				expected = []string{
					"info",
					"organizations",
					"org_quotas",
					"spaces",
					"space_quotas",
					"applications",
					"process",
					"routes",
					"route_services",
					"security_groups",
					"stacks",
					"buildpacks",
					"service_brokers",
					"service_offerings",
					"service_instances",
					"service_plans",
					"service_bindings",
					"segments",
				}
			})
			It("plans all jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("all filters are set", func() {
			BeforeEach(func() {
				active = filters.All
				expected = []string{
					"info",
					"organizations",
					"org_quotas",
					"spaces",
					"space_quotas",
					"applications",
					"process",
					"routes",
					"route_services",
					"security_groups",
					"stacks",
					"buildpacks",
					"tasks",
					"service_brokers",
					"service_offerings",
					"service_instances",
					"service_plans",
					"service_bindings",
					"segments",
					"users",
					"events",
				}
			})
			It("plans all jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("org filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.Organizations}
				expected = []string{"info", "organizations", "org_quotas"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("space filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.Spaces}
				expected = []string{"info", "spaces", "space_quotas"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("buildpack filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.Buildpacks}
				expected = []string{"info", "buildpacks"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("tasks filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.Tasks}
				expected = []string{"info", "tasks"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("isolationsegments filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.IsolationSegments}
				expected = []string{"info", "segments"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("routes filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.Routes}
				expected = []string{"info", "routes", "route_services"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("securitygroups filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.SecurityGroups}
				expected = []string{"info", "security_groups"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("servicebindings filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.ServiceBindings}
				expected = []string{"info", "service_bindings"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("serviceinstances filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.ServiceInstances}
				expected = []string{"info", "service_instances"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("services filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.Services}
				expected = []string{"info", "service_brokers", "service_offerings"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("stacks filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.Stacks}
				expected = []string{"info", "stacks"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("applications filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.Applications}
				expected = []string{"info", "organizations", "spaces", "applications", "process"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

		When("events filter is set", func() {
			BeforeEach(func() {
				active = []string{filters.Events}
				expected = []string{"info", "users", "events"}
			})
			It("plans only specific jobs", func() {
				Ω(jobs).Should(ConsistOf(expected))
			})
		})

	})
})
