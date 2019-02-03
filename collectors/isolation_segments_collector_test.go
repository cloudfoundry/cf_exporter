package collectors_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	. "github.com/bosh-prometheus/cf_exporter/collectors"
	. "github.com/bosh-prometheus/cf_exporter/utils/test_matchers"
)

func init() {
	log.Base().SetLevel("fatal")
}

var _ = Describe("IsolationSegmentsCollector", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		isolationSegmentInfoMetric                       *prometheus.GaugeVec
		isolationSegmentsScrapesTotalMetric              prometheus.Counter
		isolationSegmentsScrapeErrorsTotalMetric         prometheus.Counter
		lastIsolationSegmentsScrapeErrorMetric           prometheus.Gauge
		lastIsolationSegmentsScrapeTimestampMetric       prometheus.Gauge
		lastIsolationSegmentsScrapeDurationSecondsMetric prometheus.Gauge

		namespace             = "test_namespace"
		environment           = "test_environment"
		deployment            = "test_deployment"
		isolationSegmentId1   = "fake_isolation_segment_id_1"
		isolationSegmentName1 = "fake_isolation_segment_name_1"
		isolationSegmentId2   = "fake_isolation_segment_id_2"
		isolationSegmentName2 = "fake_isolation_segment_name_2"

		isolationSegmentsCollector *IsolationSegmentsCollector
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

		isolationSegmentInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "isolation_segment",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Isolation Segment information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"isolation_segment_id", "isolation_segment_name"},
		)
		isolationSegmentInfoMetric.WithLabelValues(
			isolationSegmentId1,
			isolationSegmentName1,
		).Set(1)
		isolationSegmentInfoMetric.WithLabelValues(
			isolationSegmentId2,
			isolationSegmentName2,
		).Set(1)

		isolationSegmentsScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "isolation_segments_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Isolation Segments.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		isolationSegmentsScrapesTotalMetric.Inc()

		isolationSegmentsScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "isolation_segments_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape error of Cloud Foundry Isolation Segments.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastIsolationSegmentsScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_isolation_segments_scrape_error",
				Help:        "Whether the last scrape of Isolation Segments metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastIsolationSegmentsScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_isolation_segments_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Isolation Segments metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastIsolationSegmentsScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_isolation_segments_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Isolation Segments metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		isolationSegmentsCollector = NewIsolationSegmentsCollector(namespace, environment, deployment, cfClient)
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
			go isolationSegmentsCollector.Describe(descriptions)
		})

		It("returns a isolation_segment_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(isolationSegmentInfoMetric.WithLabelValues(
				isolationSegmentId1,
				isolationSegmentName1,
			).Desc())))
		})

		It("returns a isolation_segments_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(isolationSegmentsScrapesTotalMetric.Desc())))
		})

		It("returns a isolation_segments_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(isolationSegmentsScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_isolation_segments_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastIsolationSegmentsScrapeErrorMetric.Desc())))
		})

		It("returns a last_isolation_segments_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastIsolationSegmentsScrapeTimestampMetric.Desc())))
		})

		It("returns a last_isolation_segments_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastIsolationSegmentsScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode                int
			isolationSegmentsResponse cfclient.ListIsolationSegmentsResponse
			metrics                   chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			isolationSegmentsResponse = cfclient.ListIsolationSegmentsResponse{
				Resources: []cfclient.IsolationSegementResponse{
					cfclient.IsolationSegementResponse{
						GUID: isolationSegmentId1,
						Name: isolationSegmentName1,
					},
					cfclient.IsolationSegementResponse{
						GUID: isolationSegmentId2,
						Name: isolationSegmentName2,
					},
				},
			}
			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/isolation_segments"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &isolationSegmentsResponse),
				),
			)

			go isolationSegmentsCollector.Collect(metrics)
		})

		It("returns a isolation_segment_info metric for isolation segment 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(isolationSegmentInfoMetric.WithLabelValues(
				isolationSegmentId1,
				isolationSegmentName1,
			))))
		})

		It("returns a isolation_segment_info metric for isolation segment 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(isolationSegmentInfoMetric.WithLabelValues(
				isolationSegmentId2,
				isolationSegmentName2,
			))))
		})

		It("returns a isolation_segments_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(isolationSegmentsScrapesTotalMetric)))
		})

		It("returns a isolation_segments_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(isolationSegmentsScrapeErrorsTotalMetric)))
		})

		It("returns a last_isolation_segments_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastIsolationSegmentsScrapeErrorMetric)))
		})

		Context("when it fails to list the isolation segments", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				isolationSegmentsScrapeErrorsTotalMetric.Inc()
				lastIsolationSegmentsScrapeErrorMetric.Set(1)
			})

			It("returns a isolation_segments_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(isolationSegmentsScrapeErrorsTotalMetric)))
			})

			It("returns a last_isolation_segments_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastIsolationSegmentsScrapeErrorMetric)))
			})
		})
	})
})
