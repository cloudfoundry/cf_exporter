package collectors_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/onsi/gomega/ghttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	. "github.com/bosh-prometheus/cf_exporter/collectors"
	. "github.com/bosh-prometheus/cf_exporter/utils/test_matchers"
)

func init() {
	log.Base().SetLevel("fatal")
}

var _ = Describe("BuildpacksCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		buildpackInfoMetric                       *prometheus.GaugeVec
		buildpacksScrapesTotalMetric              prometheus.Counter
		buildpacksScrapeErrorsTotalMetric         prometheus.Counter
		lastBuildpacksScrapeErrorMetric           prometheus.Gauge
		lastBuildpacksScrapeTimestampMetric       prometheus.Gauge
		lastBuildpacksScrapeDurationSecondsMetric prometheus.Gauge

		namespace          = "test_namespace"
		environment        = "test_environment"
		deployment         = "test_deployment"
		buildpackId1       = "fake_buildpack_id_1"
		buildpackName1     = "fake_buildpack_name_1"
		buildpackStack1    = "fake_buildpack_stack_1"
		buildpackFilename1 = "fake_buildpack_filename_1"
		buildpackId2       = "fake_buildpack_id_2"
		buildpackName2     = "fake_buildpack_name_2"
		buildpackStack2    = "fake_buildpack_stack_2"
		buildpackFilename2 = "fake_buildpack_filename_2"

		buildpacksCollector *BuildpacksCollector
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

		buildpackInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "buildpack",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Buildpack information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"buildpack_id", "buildpack_name", "buildpack_stack", "buildpack_filename"},
		)
		buildpackInfoMetric.WithLabelValues(buildpackId1, buildpackName1, buildpackStack1, buildpackFilename1).Set(1)
		buildpackInfoMetric.WithLabelValues(buildpackId2, buildpackName2, buildpackStack2, buildpackFilename2).Set(1)

		buildpacksScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "buildpacks_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Buildpacks.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		buildpacksScrapesTotalMetric.Inc()

		buildpacksScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "buildpacks_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape error of Cloud Foundry Buildpacks.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastBuildpacksScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_buildpacks_scrape_error",
				Help:        "Whether the last scrape of Buildpacks metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastBuildpacksScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_buildpacks_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Buildpacks metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastBuildpacksScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_buildpacks_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Buildpacks metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		buildpacksCollector = NewBuildpacksCollector(namespace, environment, deployment, cfClient)
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
			go buildpacksCollector.Describe(descriptions)
		})

		It("returns a buildpack_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(buildpackInfoMetric.WithLabelValues(
				buildpackId1,
				buildpackName1,
				buildpackStack1,
				buildpackFilename1,
			).Desc())))
		})

		It("returns a buildpacks_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(buildpacksScrapesTotalMetric.Desc())))
		})

		It("returns a buildpacks_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(buildpacksScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_buildpacks_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastBuildpacksScrapeErrorMetric.Desc())))
		})

		It("returns a last_buildpacks_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastBuildpacksScrapeTimestampMetric.Desc())))
		})

		It("returns a last_buildpacks_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastBuildpacksScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode         int
			buildpacksResponse cfclient.BuildpackResponse
			metrics            chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			buildpacksResponse = cfclient.BuildpackResponse{
				Resources: []cfclient.BuildpackResource{
					cfclient.BuildpackResource{
						Meta: cfclient.Meta{
							Guid: buildpackId1,
						},
						Entity: cfclient.Buildpack{
							Name:     buildpackName1,
							Filename: buildpackFilename1,
							Stack:    buildpackStack1,
						},
					},
					cfclient.BuildpackResource{
						Meta: cfclient.Meta{
							Guid: buildpackId2,
						},
						Entity: cfclient.Buildpack{
							Name:     buildpackName2,
							Filename: buildpackFilename2,
							Stack:    buildpackStack2,
						},
					},
				},
			}
			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/buildpacks"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &buildpacksResponse),
				),
			)

			go buildpacksCollector.Collect(metrics)
		})

		It("returns a buildpack_info metric for buildpack 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(buildpackInfoMetric.WithLabelValues(
				buildpackId1,
				buildpackName1,
				buildpackStack1,
				buildpackFilename1,
			))))
		})

		It("returns a buildpack_info metric for buildpack 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(buildpackInfoMetric.WithLabelValues(
				buildpackId2,
				buildpackName2,
				buildpackStack2,
				buildpackFilename2,
			))))
		})

		It("returns a buildpacks_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(buildpacksScrapesTotalMetric)))
		})

		It("returns a buildpacks_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(buildpacksScrapeErrorsTotalMetric)))
		})

		It("returns a last_buildpacks_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastBuildpacksScrapeErrorMetric)))
		})

		Context("when it fails to list the buildpacks", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				buildpacksScrapeErrorsTotalMetric.Inc()
				lastBuildpacksScrapeErrorMetric.Set(1)
			})

			It("returns a buildpacks_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(buildpacksScrapeErrorsTotalMetric)))
			})

			It("returns a last_buildpacks_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastBuildpacksScrapeErrorMetric)))
			})
		})
	})
})
