package collectors_test

import (
	"flag"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"

	. "github.com/bosh-prometheus/cf_exporter/collectors"
	. "github.com/bosh-prometheus/cf_exporter/utils/test_matchers"
)

func init() {
	flag.Set("log.level", "fatal")
}

var _ = Describe("ServicePlansCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		servicePlanInfoMetric                       *prometheus.GaugeVec
		servicePlansScrapesTotalMetric              prometheus.Counter
		servicePlansScrapeErrorsTotalMetric         prometheus.Counter
		lastServicePlansScrapeErrorMetric           prometheus.Gauge
		lastServicePlansScrapeTimestampMetric       prometheus.Gauge
		lastServicePlansScrapeDurationSecondsMetric prometheus.Gauge

		namespace        = "test_namespace"
		environment      = "test_environment"
		deployment       = "test_deployment"
		servicePlanId1   = "fake_service_plan_id_1"
		servicePlanName1 = "fake_service_plan_name_1"
		serviceId1       = "fake_service_id_1"
		servicePlanId2   = "fake_service_plan_id_2"
		servicePlanName2 = "fake_service_plan_name_2"
		serviceId2       = "fake_service_id_2"

		servicePlansCollector *ServicePlansCollector
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

		servicePlanInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "service_plan",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Service Plan information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"service_plan_id", "service_plan_name", "service_id"},
		)
		servicePlanInfoMetric.WithLabelValues(
			servicePlanId1,
			servicePlanName1,
			serviceId1,
		).Set(1)
		servicePlanInfoMetric.WithLabelValues(
			servicePlanId2,
			servicePlanName2,
			serviceId2,
		).Set(1)

		servicePlansScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "service_plans_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Service Plans.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		servicePlansScrapesTotalMetric.Inc()

		servicePlansScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "service_plans_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape error of Cloud Foundry Service Plans.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServicePlansScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_service_plans_scrape_error",
				Help:        "Whether the last scrape of Service Plans metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServicePlansScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_service_plans_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Service Plans metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServicePlansScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_service_plans_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Service Plans metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		servicePlansCollector = NewServicePlansCollector(namespace, environment, deployment, cfClient)
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
			go servicePlansCollector.Describe(descriptions)
		})

		It("returns a service_plan_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(servicePlanInfoMetric.WithLabelValues(
				servicePlanId1,
				servicePlanName1,
				serviceId1,
			).Desc())))
		})

		It("returns a service_plans_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(servicePlansScrapesTotalMetric.Desc())))
		})

		It("returns a service_plans_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(servicePlansScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_service_plans_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServicePlansScrapeErrorMetric.Desc())))
		})

		It("returns a last_service_plans_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServicePlansScrapeTimestampMetric.Desc())))
		})

		It("returns a last_service_plans_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServicePlansScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode           int
			servicePlansResponse cfclient.ServicePlansResponse
			metrics              chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			servicePlansResponse = cfclient.ServicePlansResponse{
				Resources: []cfclient.ServicePlanResource{
					cfclient.ServicePlanResource{
						Meta: cfclient.Meta{
							Guid: servicePlanId1,
						},
						Entity: cfclient.ServicePlan{
							Name:        servicePlanName1,
							ServiceGuid: serviceId1,
						},
					},
					cfclient.ServicePlanResource{
						Meta: cfclient.Meta{
							Guid: servicePlanId2,
						},
						Entity: cfclient.ServicePlan{
							Name:        servicePlanName2,
							ServiceGuid: serviceId2,
						},
					},
				},
			}
			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/service_plans"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &servicePlansResponse),
				),
			)

			go servicePlansCollector.Collect(metrics)
		})

		It("returns a service_plan_info metric for service plan 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(servicePlanInfoMetric.WithLabelValues(
				servicePlanId1,
				servicePlanName1,
				serviceId1,
			))))
		})

		It("returns a service_plan_info metric for service plan 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(servicePlanInfoMetric.WithLabelValues(
				servicePlanId2,
				servicePlanName2,
				serviceId2,
			))))
		})

		It("returns a service_plans_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(servicePlansScrapesTotalMetric)))
		})

		It("returns a service_plans_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(servicePlansScrapeErrorsTotalMetric)))
		})

		It("returns a last_service_plans_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastServicePlansScrapeErrorMetric)))
		})

		Context("when it fails to list the service plans", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				servicePlansScrapeErrorsTotalMetric.Inc()
				lastServicePlansScrapeErrorMetric.Set(1)
			})

			It("returns a service_plans_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(servicePlansScrapeErrorsTotalMetric)))
			})

			It("returns a last_service_plans_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastServicePlansScrapeErrorMetric)))
			})
		})
	})
})
