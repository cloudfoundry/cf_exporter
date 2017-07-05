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

var _ = Describe("SecurityGroupsCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		securityGroupInfoMetric                       *prometheus.GaugeVec
		securityGroupsScrapesTotalMetric              prometheus.Counter
		securityGroupsScrapeErrorsTotalMetric         prometheus.Counter
		lastSecurityGroupsScrapeErrorMetric           prometheus.Gauge
		lastSecurityGroupsScrapeTimestampMetric       prometheus.Gauge
		lastSecurityGroupsScrapeDurationSecondsMetric prometheus.Gauge

		namespace          = "test_namespace"
		environment        = "test_environment"
		deployment         = "test_deployment"
		securityGroupId1   = "fake_security_group_id_1"
		securityGroupName1 = "fake_security_group_name_1"
		securityGroupId2   = "fake_security_group_id_2"
		securityGroupName2 = "fake_security_group_name_2"

		securityGroupsCollector *SecurityGroupsCollector
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

		securityGroupInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "security_group",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Security Group information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"security_group_id", "security_group_name"},
		)
		securityGroupInfoMetric.WithLabelValues(securityGroupId1, securityGroupName1).Set(1)
		securityGroupInfoMetric.WithLabelValues(securityGroupId2, securityGroupName2).Set(1)

		securityGroupsScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "security_groups_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Security Groups.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		securityGroupsScrapesTotalMetric.Inc()

		securityGroupsScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "security_groups_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape error of Cloud Foundry Security Groups.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastSecurityGroupsScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_security_groups_scrape_error",
				Help:        "Whether the last scrape of Security Groups metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastSecurityGroupsScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_security_groups_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Security Groups metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastSecurityGroupsScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_security_groups_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Security Groups metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		securityGroupsCollector = NewSecurityGroupsCollector(namespace, environment, deployment, cfClient)
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
			go securityGroupsCollector.Describe(descriptions)
		})

		It("returns a security_group_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(securityGroupInfoMetric.WithLabelValues(
				securityGroupId1,
				securityGroupName1,
			).Desc())))
		})

		It("returns a security_groups_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(securityGroupsScrapesTotalMetric.Desc())))
		})

		It("returns a security_groups_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(securityGroupsScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_security_groups_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastSecurityGroupsScrapeErrorMetric.Desc())))
		})

		It("returns a last_security_groups_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastSecurityGroupsScrapeTimestampMetric.Desc())))
		})

		It("returns a last_security_groups_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastSecurityGroupsScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode             int
			securityGroupsResponse cfclient.SecGroupResponse
			metrics                chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			securityGroupsResponse = cfclient.SecGroupResponse{
				Resources: []cfclient.SecGroupResource{
					cfclient.SecGroupResource{
						Meta: cfclient.Meta{
							Guid: securityGroupId1,
						},
						Entity: cfclient.SecGroup{
							Name: securityGroupName1,
						},
					},
					cfclient.SecGroupResource{
						Meta: cfclient.Meta{
							Guid: securityGroupId2,
						},
						Entity: cfclient.SecGroup{
							Name: securityGroupName2,
						},
					},
				},
			}
			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/security_groups"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &securityGroupsResponse),
				),
			)

			go securityGroupsCollector.Collect(metrics)
		})

		It("returns a security_group_info metric for security group 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(securityGroupInfoMetric.WithLabelValues(
				securityGroupId1,
				securityGroupName1,
			))))
		})

		It("returns a security_group_info metric for security group 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(securityGroupInfoMetric.WithLabelValues(
				securityGroupId2,
				securityGroupName2,
			))))
		})

		It("returns a security_groups_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(securityGroupsScrapesTotalMetric)))
		})

		It("returns a security_groups_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(securityGroupsScrapeErrorsTotalMetric)))
		})

		It("returns a last_security_groups_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastSecurityGroupsScrapeErrorMetric)))
		})

		Context("when it fails to list the security groups", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				securityGroupsScrapeErrorsTotalMetric.Inc()
				lastSecurityGroupsScrapeErrorMetric.Set(1)
			})

			It("returns a security_groups_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(securityGroupsScrapeErrorsTotalMetric)))
			})

			It("returns a last_security_groups_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastSecurityGroupsScrapeErrorMetric)))
			})
		})
	})
})
