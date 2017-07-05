package collectors_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"

	. "github.com/cloudfoundry-community/cf_exporter/collectors"
	. "github.com/cloudfoundry-community/cf_exporter/utils/test_matchers"
)

var _ = Describe("RoutesCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		routeInfoMetric                       *prometheus.GaugeVec
		routesScrapesTotalMetric              prometheus.Counter
		routesScrapeErrorsTotalMetric         prometheus.Counter
		lastRoutesScrapeErrorMetric           prometheus.Gauge
		lastRoutesScrapeTimestampMetric       prometheus.Gauge
		lastRoutesScrapeDurationSecondsMetric prometheus.Gauge

		namespace               = "test_namespace"
		environment             = "test_environment"
		deployment              = "test_deployment"
		routeId1                = "fake_route_id_1"
		routeHost1              = "fake_route_host_1"
		routePath1              = "fake_route_path_1"
		routeDomainId1          = "fake_domain_id_1"
		routeSpaceId1           = "fake_space_id_1"
		routeServiceInstanceId1 = "fake_service_instance_id_1"
		routeId2                = "fake_route_id_2"
		routeHost2              = "fake_route_host_2"
		routePath2              = "fake_route_path_2"
		routeDomainId2          = "fake_domain_id_2"
		routeSpaceId2           = "fake_space_id_2"
		routeServiceInstanceId2 = "fake_service_instance_id_2"

		routesCollector *RoutesCollector
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v2/info"),
				ghttp.RespondWith(http.StatusOK, "{}"),
			),
		)

		cfClient, err = cfclient.NewClient(&cfclient.Config{
			ApiAddress: server.URL(),
			Token:      "fake-token",
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(len(server.ReceivedRequests())).To(Equal(1))

		routeInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "route",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Route information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"route_id", "route_host", "route_path", "domain_id", "space_id", "service_instance_id"},
		)
		routeInfoMetric.WithLabelValues(
			routeId1,
			routeHost1,
			routePath1,
			routeDomainId1,
			routeSpaceId1,
			routeServiceInstanceId1,
		).Set(1)
		routeInfoMetric.WithLabelValues(
			routeId2,
			routeHost2,
			routePath2,
			routeDomainId2,
			routeSpaceId2,
			routeServiceInstanceId2,
		).Set(1)

		routesScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "routes_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Routes.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		routesScrapesTotalMetric.Inc()

		routesScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "routes_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape error of Cloud Foundry Routes.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastRoutesScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_routes_scrape_error",
				Help:        "Whether the last scrape of Routes metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastRoutesScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_routes_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Routes metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastRoutesScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_routes_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Routes metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		routesCollector = NewRoutesCollector(namespace, environment, deployment, cfClient)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Describe", func() {
		var (
			descriptions chan *prometheus.Desc
		)

		BeforeEach(func() {
			descriptions = make(chan *prometheus.Desc)
		})

		JustBeforeEach(func() {
			go routesCollector.Describe(descriptions)
		})

		It("returns a route_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(routeInfoMetric.WithLabelValues(
				routeId1,
				routeHost1,
				routePath1,
				routeDomainId1,
				routeSpaceId1,
				routeServiceInstanceId1,
			).Desc())))
		})

		It("returns a routes_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(routesScrapesTotalMetric.Desc())))
		})

		It("returns a routes_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(routesScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_routes_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastRoutesScrapeErrorMetric.Desc())))
		})

		It("returns a last_routes_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastRoutesScrapeTimestampMetric.Desc())))
		})

		It("returns a last_routes_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastRoutesScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode     int
			routesResponse cfclient.RoutesResponse
			metrics        chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			routesResponse = cfclient.RoutesResponse{
				Resources: []cfclient.RoutesResource{
					cfclient.RoutesResource{
						Meta: cfclient.Meta{
							Guid: routeId1,
						},
						Entity: cfclient.Route{
							Host:                routeHost1,
							Path:                routePath1,
							DomainGuid:          routeDomainId1,
							SpaceGuid:           routeSpaceId1,
							ServiceInstanceGuid: routeServiceInstanceId1,
						},
					},
					cfclient.RoutesResource{
						Meta: cfclient.Meta{
							Guid: routeId2,
						},
						Entity: cfclient.Route{
							Host:                routeHost2,
							Path:                routePath2,
							DomainGuid:          routeDomainId2,
							SpaceGuid:           routeSpaceId2,
							ServiceInstanceGuid: routeServiceInstanceId2,
						},
					},
				},
			}
			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/routes"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &routesResponse),
				),
			)

			go routesCollector.Collect(metrics)
		})

		It("returns a route_info metric for route 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(routeInfoMetric.WithLabelValues(
				routeId1,
				routeHost1,
				routePath1,
				routeDomainId1,
				routeSpaceId1,
				routeServiceInstanceId1,
			))))
		})

		It("returns a route_info metric for route 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(routeInfoMetric.WithLabelValues(
				routeId2,
				routeHost2,
				routePath2,
				routeDomainId2,
				routeSpaceId2,
				routeServiceInstanceId2,
			))))
		})

		It("returns a routes_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(routesScrapesTotalMetric)))
		})

		It("returns a routes_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(routesScrapeErrorsTotalMetric)))
		})

		It("returns a last_routes_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastRoutesScrapeErrorMetric)))
		})

		Context("when it fails to list the routes", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				routesScrapeErrorsTotalMetric.Inc()
				lastRoutesScrapeErrorMetric.Set(1)
			})

			It("returns a routes_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(routesScrapeErrorsTotalMetric)))
			})

			It("returns a last_routes_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastRoutesScrapeErrorMetric)))
			})
		})
	})
})
