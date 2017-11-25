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

var _ = Describe("ServicesCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		serviceInfoMetric                       *prometheus.GaugeVec
		servicesScrapesTotalMetric              prometheus.Counter
		servicesScrapeErrorsTotalMetric         prometheus.Counter
		lastServicesScrapeErrorMetric           prometheus.Gauge
		lastServicesScrapeTimestampMetric       prometheus.Gauge
		lastServicesScrapeDurationSecondsMetric prometheus.Gauge

		namespace     = "test_namespace"
		environment   = "test_environment"
		deployment    = "test_deployment"
		serviceId1    = "fake_service_id_1"
		serviceLabel1 = "fake_service_label_1"
		serviceId2    = "fake_service_id_2"
		serviceLabel2 = "fake_service_label_2"

		servicesCollector *ServicesCollector
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

		serviceInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "service",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Service information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"service_id", "service_label"},
		)
		serviceInfoMetric.WithLabelValues(serviceId1, serviceLabel1).Set(1)
		serviceInfoMetric.WithLabelValues(serviceId2, serviceLabel2).Set(1)

		servicesScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "services_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Services.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		servicesScrapesTotalMetric.Inc()

		servicesScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "services_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape error of Cloud Foundry Services.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServicesScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_services_scrape_error",
				Help:        "Whether the last scrape of Services metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServicesScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_services_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Services metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServicesScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_services_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Services metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		servicesCollector = NewServicesCollector(namespace, environment, deployment, cfClient)
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
			go servicesCollector.Describe(descriptions)
		})

		It("returns a service_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(serviceInfoMetric.WithLabelValues(
				serviceId1,
				serviceLabel1,
			).Desc())))
		})

		It("returns a services_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(servicesScrapesTotalMetric.Desc())))
		})

		It("returns a services_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(servicesScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_services_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServicesScrapeErrorMetric.Desc())))
		})

		It("returns a last_services_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServicesScrapeTimestampMetric.Desc())))
		})

		It("returns a last_services_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServicesScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode       int
			servicesResponse cfclient.ServicesResponse
			metrics          chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			servicesResponse = cfclient.ServicesResponse{
				Resources: []cfclient.ServicesResource{
					cfclient.ServicesResource{
						Meta: cfclient.Meta{
							Guid: serviceId1,
						},
						Entity: cfclient.Service{
							Label: serviceLabel1,
						},
					},
					cfclient.ServicesResource{
						Meta: cfclient.Meta{
							Guid: serviceId2,
						},
						Entity: cfclient.Service{
							Label: serviceLabel2,
						},
					},
				},
			}
			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/services"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &servicesResponse),
				),
			)

			go servicesCollector.Collect(metrics)
		})

		It("returns a service_info metric for service 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(serviceInfoMetric.WithLabelValues(
				serviceId1,
				serviceLabel1,
			))))
		})

		It("returns a service_info metric for service 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(serviceInfoMetric.WithLabelValues(
				serviceId2,
				serviceLabel2,
			))))
		})

		It("returns a services_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(servicesScrapesTotalMetric)))
		})

		It("returns a services_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(servicesScrapeErrorsTotalMetric)))
		})

		It("returns a last_services_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastServicesScrapeErrorMetric)))
		})

		Context("when it fails to list the services", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				servicesScrapeErrorsTotalMetric.Inc()
				lastServicesScrapeErrorMetric.Set(1)
			})

			It("returns a services_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(servicesScrapeErrorsTotalMetric)))
			})

			It("returns a last_services_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastServicesScrapeErrorMetric)))
			})
		})
	})
})
