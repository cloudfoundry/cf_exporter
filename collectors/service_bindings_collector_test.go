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

var _ = Describe("ServiceBindingsCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		serviceBindingInfoMetric                       *prometheus.GaugeVec
		serviceBindingsScrapesTotalMetric              prometheus.Counter
		serviceBindingsScrapeErrorsTotalMetric         prometheus.Counter
		lastServiceBindingsScrapeErrorMetric           prometheus.Gauge
		lastServiceBindingsScrapeTimestampMetric       prometheus.Gauge
		lastServiceBindingsScrapeDurationSecondsMetric prometheus.Gauge

		namespace          = "test_namespace"
		environment        = "test_environment"
		deployment         = "test_deployment"
		serviceBindingId1  = "fake_service_binding_id_1"
		applicationId1     = "fake_application_id_1"
		serviceInstanceId1 = "fake_service_instance_id_1"
		serviceBindingId2  = "fake_service_binding_id_2"
		applicationId2     = "fake_application_id_2"
		serviceInstanceId2 = "fake_service_instance_id_2"

		serviceBindingsCollector *ServiceBindingsCollector
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

		serviceBindingInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "service_binding",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Service Binding information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"service_binding_id", "application_id", "service_instance_id"},
		)
		serviceBindingInfoMetric.WithLabelValues(
			serviceBindingId1,
			applicationId1,
			serviceInstanceId1,
		).Set(1)
		serviceBindingInfoMetric.WithLabelValues(
			serviceBindingId2,
			applicationId2,
			serviceInstanceId2,
		).Set(1)

		serviceBindingsScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "service_bindings_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Service Bindings.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		serviceBindingsScrapesTotalMetric.Inc()

		serviceBindingsScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "service_bindings_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape error of Cloud Foundry Service Bindings.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServiceBindingsScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_service_bindings_scrape_error",
				Help:        "Whether the last scrape of Service Bindings metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServiceBindingsScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_service_bindings_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Service Bindings metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServiceBindingsScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_service_bindings_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Service Bindings metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		serviceBindingsCollector = NewServiceBindingsCollector(namespace, environment, deployment, cfClient)
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
			go serviceBindingsCollector.Describe(descriptions)
		})

		It("returns a service_binding_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(serviceBindingInfoMetric.WithLabelValues(
				serviceBindingId1,
				applicationId1,
				serviceInstanceId1,
			).Desc())))
		})

		It("returns a service_bindings_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(serviceBindingsScrapesTotalMetric.Desc())))
		})

		It("returns a service_bindings_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(serviceBindingsScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_service_bindings_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServiceBindingsScrapeErrorMetric.Desc())))
		})

		It("returns a last_service_bindings_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServiceBindingsScrapeTimestampMetric.Desc())))
		})

		It("returns a last_service_bindings_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServiceBindingsScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode              int
			serviceBindingsResponse cfclient.ServiceBindingsResponse
			metrics                 chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			serviceBindingsResponse = cfclient.ServiceBindingsResponse{
				Resources: []cfclient.ServiceBindingResource{
					cfclient.ServiceBindingResource{
						Meta: cfclient.Meta{
							Guid: serviceBindingId1,
						},
						Entity: cfclient.ServiceBinding{
							AppGuid:             applicationId1,
							ServiceInstanceGuid: serviceInstanceId1,
						},
					},
					cfclient.ServiceBindingResource{
						Meta: cfclient.Meta{
							Guid: serviceBindingId2,
						},
						Entity: cfclient.ServiceBinding{
							AppGuid:             applicationId2,
							ServiceInstanceGuid: serviceInstanceId2,
						},
					},
				},
			}
			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/service_bindings"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &serviceBindingsResponse),
				),
			)

			go serviceBindingsCollector.Collect(metrics)
		})

		It("returns a service_binding_info metric for service binding 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(serviceBindingInfoMetric.WithLabelValues(
				serviceBindingId1,
				applicationId1,
				serviceInstanceId1,
			))))
		})

		It("returns a service_binding_info metric for service binding 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(serviceBindingInfoMetric.WithLabelValues(
				serviceBindingId2,
				applicationId2,
				serviceInstanceId2,
			))))
		})

		It("returns a service_bindings_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(serviceBindingsScrapesTotalMetric)))
		})

		It("returns a service_bindings_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(serviceBindingsScrapeErrorsTotalMetric)))
		})

		It("returns a last_service_bindings_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastServiceBindingsScrapeErrorMetric)))
		})

		Context("when it fails to list the service bindings", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				serviceBindingsScrapeErrorsTotalMetric.Inc()
				lastServiceBindingsScrapeErrorMetric.Set(1)
			})

			It("returns a service_bindings_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(serviceBindingsScrapeErrorsTotalMetric)))
			})

			It("returns a last_service_bindings_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastServiceBindingsScrapeErrorMetric)))
			})
		})
	})
})
