package fetcher

import (
	"fmt"
	"net"
	"net/http"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/resources"
	"github.com/cloudfoundry/cf_exporter/v2/models"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry/cf_exporter/v2/filters"
)

var _ = ginkgo.Describe("Fetcher", func() {
	ginkgo.Context("fetching jobs are planned according to filter", func() {
		var (
			fetcher  *Fetcher
			active   []string
			jobs     []string
			expected []string
		)

		ginkgo.JustBeforeEach(func() {
			f, err := filters.NewFilter(active...)
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
			fetcher = NewFetcher(10, &CFConfig{}, &BBSConfig{}, f)
			gomega.Ω(fetcher).ShouldNot(gomega.BeNil())
			fetcher.workInit()

			close(fetcher.worker.list)
			jobs = []string{}
			for w := range fetcher.worker.list {
				jobs = append(jobs, w.name)
			}
		})

		ginkgo.When("default filters are set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{}
				expected = []string{
					"info",
					"organizations",
					"org_quotas",
					"spaces",
					"space_quotas",
					"applications",
					"droplets",
					"domains",
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
					"service_route_bindings",
					"segments",
					"actual_lrps",
				}
			})
			ginkgo.It("plans all jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("all filters are set", func() {
			ginkgo.BeforeEach(func() {
				active = filters.All
				expected = []string{
					"info",
					"organizations",
					"org_quotas",
					"spaces",
					"space_quotas",
					"applications",
					"droplets",
					"domains",
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
					"service_route_bindings",
					"segments",
					"users",
					"events",
					"actual_lrps",
				}
			})
			ginkgo.It("plans all jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("org filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Organizations}
				expected = []string{"info", "organizations", "org_quotas"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("space filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Spaces}
				expected = []string{"info", "spaces", "space_quotas"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("buildpack filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Buildpacks}
				expected = []string{"info", "buildpacks"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("tasks filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Tasks}
				expected = []string{"info", "tasks"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("isolationsegments filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.IsolationSegments}
				expected = []string{"info", "segments"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("routes filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Routes}
				expected = []string{"info", "routes", "route_services"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("securitygroups filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.SecurityGroups}
				expected = []string{"info", "security_groups"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("servicebindings filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.ServiceBindings}
				expected = []string{"info", "service_bindings"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("serviceinstances filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.ServiceInstances}
				expected = []string{"info", "service_instances"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("services filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Services}
				expected = []string{"info", "service_brokers", "service_offerings"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("stacks filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Stacks}
				expected = []string{"info", "stacks"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("applications filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Applications}
				expected = []string{"info", "organizations", "spaces", "applications", "process"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("events filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Events}
				expected = []string{"info", "users", "events"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("droplets filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Droplets}
				expected = []string{"info", "droplets"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("actual_lrps filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.ActualLRPs}
				expected = []string{"info", "actual_lrps"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

	})

	ginkgo.Context("disabling filters during a scrape", func() {
		ginkgo.It("does not mutate the original filter used for future scrapes", func() {
			filter, err := filters.NewFilter()
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())

			fetcher := NewFetcher(10, &CFConfig{}, &BBSConfig{}, filter)
			fetcher.filters.Disable([]string{filters.ActualLRPs})

			gomega.Ω(fetcher.filters.Enabled(filters.ActualLRPs)).Should(gomega.BeFalse())
			gomega.Ω(filter.Enabled(filters.ActualLRPs)).Should(gomega.BeTrue())
		})
	})

	ginkgo.Context("when BBS client initialization fails", func() {
		var server *ghttp.Server

		ginkgo.BeforeEach(func() {
			tokenResponse := fmt.Sprintf(`{"access_token": "%s", "refresh_token": "value"}`, fakeToken)
			server = ghttp.NewServer()

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/"),
					ghttp.RespondWith(http.StatusOK, serialize(ccv3.Root{
						Links: ccv3.RootLinks{
							Login: resources.APILink{HREF: server.URL()},
							UAA:   resources.APILink{HREF: server.URL()},
						},
					})),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token"),
					ghttp.RespondWith(http.StatusOK, tokenResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token"),
					ghttp.RespondWith(http.StatusOK, tokenResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token"),
					ghttp.RespondWith(http.StatusOK, tokenResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/info"),
					ghttp.RespondWith(http.StatusOK, serialize(models.Info{Name: "test-foundation"})),
				),
			)
		})

		ginkgo.AfterEach(func() {
			server.Close()
		})

		ginkgo.It("disables actual lrps only for the current scrape", func() {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
			refusedURL := fmt.Sprintf("http://%s", listener.Addr().String())
			gomega.Ω(listener.Close()).Should(gomega.Succeed())

			filter, err := filters.NewFilter(filters.ActualLRPs)
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())

			fetcher := NewFetcher(1, &CFConfig{
				URL:          server.URL(),
				ClientID:     "fake",
				ClientSecret: "fake",
			}, &BBSConfig{
				URL:     refusedURL,
				Timeout: 1,
			}, filter)

			objs := fetcher.GetObjects()

			gomega.Ω(objs.Error).ShouldNot(gomega.HaveOccurred())
			gomega.Ω(objs.BBSActualLRPsError).Should(gomega.HaveOccurred())
			gomega.Ω(objs.Info.Name).Should(gomega.Equal("test-foundation"))
			gomega.Ω(objs.ProcessActualLRPs).Should(gomega.BeEmpty())
			gomega.Ω(fetcher.filters.Enabled(filters.ActualLRPs)).Should(gomega.BeFalse())
			gomega.Ω(filter.Enabled(filters.ActualLRPs)).Should(gomega.BeTrue())
		})
	})
})
