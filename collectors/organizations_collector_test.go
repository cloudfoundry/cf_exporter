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

var _ = Describe("OrganizationsCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		organizationInfoMetric                         *prometheus.GaugeVec
		organizationNonBasicServicesAllowedMetric      *prometheus.GaugeVec
		organizationInstanceMemoryMbLimitMetric        *prometheus.GaugeVec
		organizationTotalAppInstancesQuotaMetric       *prometheus.GaugeVec
		organizationTotalAppTasksQuotaMetric           *prometheus.GaugeVec
		organizationTotalMemoryMbQuotaMetric           *prometheus.GaugeVec
		organizationTotalPrivateDomainsQuotaMetric     *prometheus.GaugeVec
		organizationTotalReservedRoutePortsQuotaMetric *prometheus.GaugeVec
		organizationTotalRoutesQuotaMetric             *prometheus.GaugeVec
		organizationTotalServiceKeysQuotaMetric        *prometheus.GaugeVec
		organizationTotalServicesQuotaMetric           *prometheus.GaugeVec
		organizationsScrapesTotalMetric                prometheus.Counter
		organizationsScrapeErrorsTotalMetric           prometheus.Counter
		lastOrganizationsScrapeErrorMetric             prometheus.Gauge
		lastOrganizationsScrapeTimestampMetric         prometheus.Gauge
		lastOrganizationsScrapeDurationSecondsMetric   prometheus.Gauge

		namespace   = "test_namespace"
		environment = "test_environment"
		deployment  = "test_deployment"

		organizationId1          = "fake_organization_id_1"
		organizationName1        = "fake_organization_name_1"
		quotaDefinitionId1       = "fake_quota_definition_id_1"
		quotaDefinitionName1     = "fake_quota_definition_name_1"
		NonBasicServicesAllowed1 = 0
		InstanceMemoryLimit1     = 11
		AppInstanceLimit1        = 12
		AppTaskLimit1            = 13
		MemoryLimit1             = 14
		TotalPrivateDomains1     = 15
		TotalReservedRoutePorts1 = 16
		TotalRoutes1             = 17
		TotalServiceKeys1        = 18
		TotalServices1           = 19

		organizationId2          = "fake_organization_id_2"
		organizationName2        = "fake_organization_name_2"
		quotaDefinitionId2       = "fake_quota_definition_id_2"
		quotaDefinitionName2     = "fake_quota_definition_name_2"
		NonBasicServicesAllowed2 = 1
		InstanceMemoryLimit2     = 21
		AppInstanceLimit2        = 22
		AppTaskLimit2            = 23
		MemoryLimit2             = 24
		TotalPrivateDomains2     = 25
		TotalReservedRoutePorts2 = 26
		TotalRoutes2             = 27
		TotalServiceKeys2        = 28
		TotalServices2           = 29

		organizationId3      = "fake_organization_id_3"
		organizationName3    = "fake_organization_name_3"
		quotaDefinitionId3   = ""
		quotaDefinitionName3 = ""

		organizationsCollector *OrganizationsCollector
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

		organizationInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Organization information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name", "quota_name"},
		)
		organizationInfoMetric.WithLabelValues(organizationId1, organizationName1, quotaDefinitionName1).Set(float64(1))
		organizationInfoMetric.WithLabelValues(organizationId2, organizationName2, quotaDefinitionName2).Set(float64(1))
		organizationInfoMetric.WithLabelValues(organizationId3, organizationName3, quotaDefinitionName3).Set(float64(1))

		organizationNonBasicServicesAllowedMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "non_basic_services_allowed",
				Help:        "A Cloud Foundry Organization can provision instances of paid service plans? (1 for true, 0 for false).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name"},
		)
		organizationNonBasicServicesAllowedMetric.WithLabelValues(organizationId1, organizationName1).Set(float64(NonBasicServicesAllowed1))
		organizationNonBasicServicesAllowedMetric.WithLabelValues(organizationId2, organizationName2).Set(float64(NonBasicServicesAllowed2))

		organizationInstanceMemoryMbLimitMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "instance_memory_mb_limit",
				Help:        "Maximum amount of memory (Mb) an application instance can have in a Cloud Foundry Organization.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name"},
		)
		organizationInstanceMemoryMbLimitMetric.WithLabelValues(organizationId1, organizationName1).Set(float64(InstanceMemoryLimit1))
		organizationInstanceMemoryMbLimitMetric.WithLabelValues(organizationId2, organizationName2).Set(float64(InstanceMemoryLimit2))

		organizationTotalAppInstancesQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "total_app_instances_quota",
				Help:        "Total number of application instances that may be created in a Cloud Foundry Organization.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name"},
		)
		organizationTotalAppInstancesQuotaMetric.WithLabelValues(organizationId1, organizationName1).Set(float64(AppInstanceLimit1))
		organizationTotalAppInstancesQuotaMetric.WithLabelValues(organizationId2, organizationName2).Set(float64(AppInstanceLimit2))

		organizationTotalAppTasksQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "total_app_tasks_quota",
				Help:        "Total number of application tasks that may be created in a Cloud Foundry Organization.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name"},
		)
		organizationTotalAppTasksQuotaMetric.WithLabelValues(organizationId1, organizationName1).Set(float64(AppTaskLimit1))
		organizationTotalAppTasksQuotaMetric.WithLabelValues(organizationId2, organizationName2).Set(float64(AppTaskLimit2))

		organizationTotalMemoryMbQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "total_memory_mb_quota",
				Help:        "Total amount of memory (Mb) a Cloud Foundry Organization can have.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name"},
		)
		organizationTotalMemoryMbQuotaMetric.WithLabelValues(organizationId1, organizationName1).Set(float64(MemoryLimit1))
		organizationTotalMemoryMbQuotaMetric.WithLabelValues(organizationId2, organizationName2).Set(float64(MemoryLimit2))

		organizationTotalPrivateDomainsQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "total_private_domains_quota",
				Help:        "Total number of private domains that may be created in a Cloud Foundry Organization.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name"},
		)
		organizationTotalPrivateDomainsQuotaMetric.WithLabelValues(organizationId1, organizationName1).Set(float64(MemoryLimit1))
		organizationTotalPrivateDomainsQuotaMetric.WithLabelValues(organizationId2, organizationName2).Set(float64(MemoryLimit2))

		organizationTotalReservedRoutePortsQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "total_reserved_route_ports_quota",
				Help:        "Total number of routes that may be created with reserved ports in a Cloud Foundry Organization.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name"},
		)
		organizationTotalReservedRoutePortsQuotaMetric.WithLabelValues(organizationId1, organizationName1).Set(float64(TotalReservedRoutePorts1))
		organizationTotalReservedRoutePortsQuotaMetric.WithLabelValues(organizationId2, organizationName2).Set(float64(TotalReservedRoutePorts2))

		organizationTotalRoutesQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "total_routes_quota",
				Help:        "Total number of routes that may be created in a Cloud Foundry Organization.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name"},
		)
		organizationTotalRoutesQuotaMetric.WithLabelValues(organizationId1, organizationName1).Set(float64(TotalRoutes1))
		organizationTotalRoutesQuotaMetric.WithLabelValues(organizationId2, organizationName2).Set(float64(TotalRoutes2))

		organizationTotalServiceKeysQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "total_service_keys_quota",
				Help:        "Total number of service keys that may be created in a Cloud Foundry Organization.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name"},
		)
		organizationTotalServiceKeysQuotaMetric.WithLabelValues(organizationId1, organizationName1).Set(float64(TotalServiceKeys1))
		organizationTotalServiceKeysQuotaMetric.WithLabelValues(organizationId2, organizationName2).Set(float64(TotalServiceKeys2))

		organizationTotalServicesQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "organization",
				Name:        "total_services_quota",
				Help:        "Total number of service instances that may be created in a Cloud Foundry Organization.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"organization_id", "organization_name"},
		)
		organizationTotalServicesQuotaMetric.WithLabelValues(organizationId1, organizationName1).Set(float64(TotalServices1))
		organizationTotalServicesQuotaMetric.WithLabelValues(organizationId2, organizationName2).Set(float64(TotalServices2))

		organizationsScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "organizations_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Organizations.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		organizationsScrapesTotalMetric.Inc()

		organizationsScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "organizations_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape errors of Cloud Foundry Organizations.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastOrganizationsScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_organizations_scrape_error",
				Help:        "Whether the last scrape of Organizations metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastOrganizationsScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_organizations_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Organizations metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastOrganizationsScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_organizations_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Organizations metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		organizationsCollector = NewOrganizationsCollector(namespace, environment, deployment, cfClient)
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
			go organizationsCollector.Describe(descriptions)
		})

		It("returns a organization_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationInfoMetric.WithLabelValues(
				organizationId1,
				organizationName1,
				quotaDefinitionName1,
			).Desc())))
		})

		It("returns a organization_non_basic_services_allowed metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationNonBasicServicesAllowedMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			).Desc())))
		})

		It("returns a organization_instance_memory_mb_limit metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationInstanceMemoryMbLimitMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			).Desc())))
		})

		It("returns a organization_total_app_instances_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationTotalAppInstancesQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			).Desc())))
		})

		It("returns a organization_total_app_tasks_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationTotalAppTasksQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			).Desc())))
		})

		It("returns a organization_total_memory_mb_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationTotalMemoryMbQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			).Desc())))
		})

		It("returns a organization_total_private_domains_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationTotalPrivateDomainsQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			).Desc())))
		})

		It("returns a organization_total_reserved_route_ports_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationTotalReservedRoutePortsQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			).Desc())))
		})

		It("returns a organization_total_routes_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationTotalRoutesQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			).Desc())))
		})

		It("returns a organization_total_service_keys_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationTotalServiceKeysQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			).Desc())))
		})

		It("returns a organization_total_services_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationTotalServicesQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			).Desc())))
		})

		It("returns a organizations_scrapes metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationsScrapesTotalMetric.Desc())))
		})

		It("returns a organizations_scrape_errors metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(organizationsScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_organizations_scrape_errorr metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastOrganizationsScrapeErrorMetric.Desc())))
		})

		It("returns a last_organizations_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastOrganizationsScrapeTimestampMetric.Desc())))
		})

		It("returns a last_organizations_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastOrganizationsScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode        int
			orgQuotasResponse cfclient.OrgQuotasResponse
			orgsResponse      cfclient.OrgResponse
			metrics           chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			orgQuotasResponse = cfclient.OrgQuotasResponse{
				Resources: []cfclient.OrgQuotasResource{
					cfclient.OrgQuotasResource{
						Meta: cfclient.Meta{
							Guid: quotaDefinitionId1,
						},
						Entity: cfclient.OrgQuota{
							Name: quotaDefinitionName1,
							NonBasicServicesAllowed: false,
							TotalServices:           TotalServices1,
							TotalRoutes:             TotalRoutes1,
							TotalPrivateDomains:     TotalPrivateDomains1,
							MemoryLimit:             MemoryLimit1,
							InstanceMemoryLimit:     InstanceMemoryLimit1,
							AppInstanceLimit:        AppInstanceLimit1,
							AppTaskLimit:            AppTaskLimit1,
							TotalServiceKeys:        TotalServiceKeys1,
							TotalReservedRoutePorts: TotalReservedRoutePorts1,
						},
					},
					cfclient.OrgQuotasResource{
						Meta: cfclient.Meta{
							Guid: quotaDefinitionId2,
						},
						Entity: cfclient.OrgQuota{
							Name: quotaDefinitionName2,
							NonBasicServicesAllowed: true,
							TotalServices:           TotalServices2,
							TotalRoutes:             TotalRoutes2,
							TotalPrivateDomains:     TotalPrivateDomains2,
							MemoryLimit:             MemoryLimit2,
							InstanceMemoryLimit:     InstanceMemoryLimit2,
							AppInstanceLimit:        AppInstanceLimit2,
							AppTaskLimit:            AppTaskLimit2,
							TotalServiceKeys:        TotalServiceKeys2,
							TotalReservedRoutePorts: TotalReservedRoutePorts2,
						},
					},
				},
			}

			orgsResponse = cfclient.OrgResponse{
				Resources: []cfclient.OrgResource{
					cfclient.OrgResource{
						Meta: cfclient.Meta{
							Guid: organizationId1,
						},
						Entity: cfclient.Org{
							Name:                organizationName1,
							QuotaDefinitionGuid: quotaDefinitionId1,
						},
					},
					cfclient.OrgResource{
						Meta: cfclient.Meta{
							Guid: organizationId2,
						},
						Entity: cfclient.Org{
							Name:                organizationName2,
							QuotaDefinitionGuid: quotaDefinitionId2,
						},
					},
					cfclient.OrgResource{
						Meta: cfclient.Meta{
							Guid: organizationId3,
						},
						Entity: cfclient.Org{
							Name:                organizationName3,
							QuotaDefinitionGuid: quotaDefinitionId3,
						},
					},
				},
			}
			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/quota_definitions"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &orgQuotasResponse),
				),
				ghttp.CombineHandlers(

					ghttp.VerifyRequest("GET", "/v2/organizations"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &orgsResponse),
				),
			)

			go organizationsCollector.Collect(metrics)
		})

		It("returns a organization_info metric for organization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationInfoMetric.WithLabelValues(
				organizationId1,
				organizationName1,
				quotaDefinitionName1,
			))))
		})

		It("returns a organization_info metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationInfoMetric.WithLabelValues(
				organizationId2,
				organizationName2,
				quotaDefinitionName2,
			))))
		})

		It("returns a organization_info metric for organization 3", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationInfoMetric.WithLabelValues(
				organizationId3,
				organizationName3,
				quotaDefinitionName3,
			))))
		})

		It("returns a organization_non_basic_services_allowed metric fororganization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationNonBasicServicesAllowedMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			))))
		})

		It("returns a organization_non_basic_services_allowed metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationNonBasicServicesAllowedMetric.WithLabelValues(
				organizationId2,
				organizationName2,
			))))
		})

		It("does not returns a organization_non_basic_services_allowed metric for organization 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(organizationNonBasicServicesAllowedMetric.WithLabelValues(
				organizationId3,
				organizationName3,
			))))
		})

		It("returns a organization_instance_memory_mb_limit metric for organization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationInstanceMemoryMbLimitMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			))))
		})

		It("returns a organization_instance_memory_mb_limit metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationInstanceMemoryMbLimitMetric.WithLabelValues(
				organizationId2,
				organizationName2,
			))))
		})

		It("does not returns a organization_instance_memory_mb_limit metric for organization 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(organizationInstanceMemoryMbLimitMetric.WithLabelValues(
				organizationId3,
				organizationName3,
			))))
		})

		It("returns a organization_total_app_instances_quota metric for organization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalAppInstancesQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			))))
		})

		It("returns a organization_total_app_instances_quota metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalAppInstancesQuotaMetric.WithLabelValues(
				organizationId2,
				organizationName2,
			))))
		})

		It("does not returns a organizatione_total_app_instances_quota metric for organization 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(organizationTotalAppInstancesQuotaMetric.WithLabelValues(
				organizationId3,
				organizationName3,
			))))
		})

		It("returns a organization_total_app_tasks_quota metric for organization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalAppTasksQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			))))
		})

		It("returns a organization_total_app_tasks_quota metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalAppTasksQuotaMetric.WithLabelValues(
				organizationId2,
				organizationName2,
			))))
		})

		It("does not returns a organization_total_app_tasks_quota metric for organization 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(organizationTotalAppTasksQuotaMetric.WithLabelValues(
				organizationId3,
				organizationName3,
			))))
		})

		It("returns a organization_total_memory_mb_quota metric for organization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalMemoryMbQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			))))
		})

		It("returns a organization_total_memory_mb_quota metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalMemoryMbQuotaMetric.WithLabelValues(
				organizationId2,
				organizationName2,
			))))
		})

		It("does not returns a organization_total_memory_mb_quota metric for organization 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(organizationTotalMemoryMbQuotaMetric.WithLabelValues(
				organizationId3,
				organizationName3,
			))))
		})

		It("returns a organization_total_private_domains_quota metric for organization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalPrivateDomainsQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			))))
		})

		It("returns a organization_total_private_domains_quota metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalPrivateDomainsQuotaMetric.WithLabelValues(
				organizationId2,
				organizationName2,
			))))
		})

		It("does not returns a organization_total_private_domains_quota metric for organization 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(organizationTotalPrivateDomainsQuotaMetric.WithLabelValues(
				organizationId3,
				organizationName3,
			))))
		})

		It("returns a organization_total_reserved_route_ports_quota metric for organization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalReservedRoutePortsQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			))))
		})

		It("returns a organization_total_reserved_route_ports_quota metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalReservedRoutePortsQuotaMetric.WithLabelValues(
				organizationId2,
				organizationName2,
			))))
		})

		It("does not returns a sorganization_total_reserved_route_ports_quota metric for organization 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(organizationTotalReservedRoutePortsQuotaMetric.WithLabelValues(
				organizationId3,
				organizationName3,
			))))
		})

		It("returns a sorganization_total_routes_quota metric for organization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalRoutesQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			))))
		})

		It("returns a organization_total_routes_quota metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalRoutesQuotaMetric.WithLabelValues(
				organizationId2,
				organizationName2,
			))))
		})

		It("does not returns a organization_total_routes_quota metric for organization 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(organizationTotalRoutesQuotaMetric.WithLabelValues(
				organizationId3,
				organizationName3,
			))))
		})

		It("returns a organization_total_service_keys_quota metric for organization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalServiceKeysQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			))))
		})

		It("returns a organization_total_service_keys_quota metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalServiceKeysQuotaMetric.WithLabelValues(
				organizationId2,
				organizationName2,
			))))
		})

		It("does not returns a organization_total_service_keys_quota metric for organization 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(organizationTotalServiceKeysQuotaMetric.WithLabelValues(
				organizationId3,
				organizationName3,
			))))
		})

		It("returns a organization_total_services_quota metric for organization 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalServicesQuotaMetric.WithLabelValues(
				organizationId1,
				organizationName1,
			))))
		})

		It("returns a organization_total_services_quota metric for organization 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationTotalServicesQuotaMetric.WithLabelValues(
				organizationId2,
				organizationName2,
			))))
		})

		It("does not returns a organization_total_services_quota metric for organization 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(organizationTotalServicesQuotaMetric.WithLabelValues(
				organizationId3,
				organizationName3,
			))))
		})

		It("returns a organizations_scrapes metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationsScrapesTotalMetric)))
		})

		It("returns a organizations_scrape_errors metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(organizationsScrapeErrorsTotalMetric)))
		})

		It("returns a ast_organizations_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastOrganizationsScrapeErrorMetric)))
		})

		Context("when it fails to list the organizations", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				organizationsScrapeErrorsTotalMetric.Inc()
				lastOrganizationsScrapeErrorMetric.Set(1)
			})

			It("returns a organizations_scrape_errors metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(organizationsScrapeErrorsTotalMetric)))
			})

			It("returns a ast_organizations_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastOrganizationsScrapeErrorMetric)))
			})
		})
	})
})
