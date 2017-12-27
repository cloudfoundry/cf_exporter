package collectors_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	. "github.com/bosh-prometheus/cf_exporter/collectors"
	. "github.com/bosh-prometheus/cf_exporter/utils/test_matchers"
)

func init() {
	log.Base().SetLevel("fatal")
}

var _ = Describe("StacksCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		stackInfoMetric                       *prometheus.GaugeVec
		stacksScrapesTotalMetric              prometheus.Counter
		stacksScrapeErrorsTotalMetric         prometheus.Counter
		lastStacksScrapeErrorMetric           prometheus.Gauge
		lastStacksScrapeTimestampMetric       prometheus.Gauge
		lastStacksScrapeDurationSecondsMetric prometheus.Gauge

		namespace   = "test_namespace"
		environment = "test_environment"
		deployment  = "test_deployment"
		stackId1    = "fake_stack_id_1"
		stackName1  = "fake_stack_name_1"
		stackId2    = "fake_stack_id_2"
		stackName2  = "fake_stack_name_2"

		stacksCollector *StacksCollector
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

		stackInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "stack",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Stack information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"stack_id", "stack_name"},
		)
		stackInfoMetric.WithLabelValues(stackId1, stackName1).Set(1)
		stackInfoMetric.WithLabelValues(stackId2, stackName2).Set(1)

		stacksScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "stacks_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Stacks.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		stacksScrapesTotalMetric.Inc()

		stacksScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "stacks_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape error of Cloud Foundry Stacks.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastStacksScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_stacks_scrape_error",
				Help:        "Whether the last scrape of Stacks metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastStacksScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_stacks_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Stacks metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastStacksScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_stacks_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Stacks metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		stacksCollector = NewStacksCollector(namespace, environment, deployment, cfClient)
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
			go stacksCollector.Describe(descriptions)
		})

		It("returns a stack_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(stackInfoMetric.WithLabelValues(
				stackId1,
				stackName1,
			).Desc())))
		})

		It("returns a stacks_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(stacksScrapesTotalMetric.Desc())))
		})

		It("returns a stacks_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(stacksScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_stacks_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastStacksScrapeErrorMetric.Desc())))
		})

		It("returns a last_stacks_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastStacksScrapeTimestampMetric.Desc())))
		})

		It("returns a last_stacks_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastStacksScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode     int
			stacksResponse cfclient.StacksResponse
			metrics        chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			stacksResponse = cfclient.StacksResponse{
				Resources: []cfclient.StacksResource{
					cfclient.StacksResource{
						Meta: cfclient.Meta{
							Guid: stackId1,
						},
						Entity: cfclient.Stack{
							Name: stackName1,
						},
					},
					cfclient.StacksResource{
						Meta: cfclient.Meta{
							Guid: stackId2,
						},
						Entity: cfclient.Stack{
							Name: stackName2,
						},
					},
				},
			}
			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/stacks"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &stacksResponse),
				),
			)

			go stacksCollector.Collect(metrics)
		})

		It("returns a stack_info metric for stack 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(stackInfoMetric.WithLabelValues(
				stackId1,
				stackName1,
			))))
		})

		It("returns a stack_info metric for stack 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(stackInfoMetric.WithLabelValues(
				stackId2,
				stackName2,
			))))
		})

		It("returns a stacks_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(stacksScrapesTotalMetric)))
		})

		It("returns a stacks_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(stacksScrapeErrorsTotalMetric)))
		})

		It("returns a last_stacks_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastStacksScrapeErrorMetric)))
		})

		Context("when it fails to list the stacks", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				stacksScrapeErrorsTotalMetric.Inc()
				lastStacksScrapeErrorMetric.Set(1)
			})

			It("returns a stacks_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(stacksScrapeErrorsTotalMetric)))
			})

			It("returns a last_stacks_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastStacksScrapeErrorMetric)))
			})
		})
	})
})
