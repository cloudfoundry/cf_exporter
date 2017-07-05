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

var _ = Describe("SpacesCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		spaceInfoMetric                         *prometheus.GaugeVec
		spaceNonBasicServicesAllowedMetric      *prometheus.GaugeVec
		spaceInstanceMemoryMbLimitMetric        *prometheus.GaugeVec
		spaceTotalAppInstancesQuotaMetric       *prometheus.GaugeVec
		spaceTotalAppTasksQuotaMetric           *prometheus.GaugeVec
		spaceTotalMemoryMbQuotaMetric           *prometheus.GaugeVec
		spaceTotalReservedRoutePortsQuotaMetric *prometheus.GaugeVec
		spaceTotalRoutesQuotaMetric             *prometheus.GaugeVec
		spaceTotalServiceKeysQuotaMetric        *prometheus.GaugeVec
		spaceTotalServicesQuotaMetric           *prometheus.GaugeVec
		spacesScrapesTotalMetric                prometheus.Counter
		spacesScrapeErrorsTotalMetric           prometheus.Counter
		lastSpacesScrapeErrorMetric             prometheus.Gauge
		lastSpacesScrapeTimestampMetric         prometheus.Gauge
		lastSpacesScrapeDurationSecondsMetric   prometheus.Gauge

		namespace   = "test_namespace"
		environment = "test_environment"
		deployment  = "test_deployment"

		spaceId1                 = "fake_space_id_1"
		spaceName1               = "fake_space_name_1"
		organizationId1          = "fake_organization_id_1"
		quotaDefinitionId1       = "fake_quota_definition_id_1"
		quotaDefinitionName1     = "fake_quota_definition_name_1"
		NonBasicServicesAllowed1 = 0
		InstanceMemoryLimit1     = 11
		AppInstanceLimit1        = 12
		AppTaskLimit1            = 13
		MemoryLimit1             = 14
		TotalReservedRoutePorts1 = 15
		TotalRoutes1             = 16
		TotalServiceKeys1        = 17
		TotalServices1           = 18

		spaceId2                 = "fake_space_id_2"
		spaceName2               = "fake_space_name_2"
		organizationId2          = "fake_organization_id_2"
		quotaDefinitionId2       = "fake_quota_definition_id_2"
		quotaDefinitionName2     = "fake_quota_definition_name_2"
		NonBasicServicesAllowed2 = 1
		InstanceMemoryLimit2     = 21
		AppInstanceLimit2        = 22
		AppTaskLimit2            = 23
		MemoryLimit2             = 24
		TotalReservedRoutePorts2 = 25
		TotalRoutes2             = 26
		TotalServiceKeys2        = 27
		TotalServices2           = 28

		spaceId3             = "fake_space_id_3"
		spaceName3           = "fake_space_name_3"
		organizationId3      = "fake_organization_id_3"
		quotaDefinitionId3   = ""
		quotaDefinitionName3 = ""

		spacesCollector *SpacesCollector
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

		spaceInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "space",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Space information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"space_id", "space_name", "organization_id", "quota_name"},
		)
		spaceInfoMetric.WithLabelValues(spaceId1, spaceName1, organizationId1, quotaDefinitionName1).Set(float64(1))
		spaceInfoMetric.WithLabelValues(spaceId2, spaceName2, organizationId2, quotaDefinitionName2).Set(float64(1))
		spaceInfoMetric.WithLabelValues(spaceId3, spaceName3, organizationId3, quotaDefinitionName3).Set(float64(1))

		spaceNonBasicServicesAllowedMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "space",
				Name:        "non_basic_services_allowed",
				Help:        "A Cloud Foundry Space can provision instances of paid service plans? (1 for true, 0 for false).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"space_id", "space_name", "organization_id"},
		)
		spaceNonBasicServicesAllowedMetric.WithLabelValues(spaceId1, spaceName1, organizationId1).Set(float64(NonBasicServicesAllowed1))
		spaceNonBasicServicesAllowedMetric.WithLabelValues(spaceId2, spaceName2, organizationId2).Set(float64(NonBasicServicesAllowed2))

		spaceInstanceMemoryMbLimitMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "space",
				Name:        "instance_memory_mb_limit",
				Help:        "Maximum amount of memory (Mb) an application instance can have in a Cloud Foundry Space.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"space_id", "space_name", "organization_id"},
		)
		spaceInstanceMemoryMbLimitMetric.WithLabelValues(spaceId1, spaceName1, organizationId1).Set(float64(InstanceMemoryLimit1))
		spaceInstanceMemoryMbLimitMetric.WithLabelValues(spaceId2, spaceName2, organizationId2).Set(float64(InstanceMemoryLimit2))

		spaceTotalAppInstancesQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "space",
				Name:        "total_app_instances_quota",
				Help:        "Total number of application instances that may be created in a Cloud Foundry Space.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"space_id", "space_name", "organization_id"},
		)
		spaceTotalAppInstancesQuotaMetric.WithLabelValues(spaceId1, spaceName1, organizationId1).Set(float64(AppInstanceLimit1))
		spaceTotalAppInstancesQuotaMetric.WithLabelValues(spaceId2, spaceName2, organizationId2).Set(float64(AppInstanceLimit2))

		spaceTotalAppTasksQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "space",
				Name:        "total_app_tasks_quota",
				Help:        "Total number of application tasks that may be created in a Cloud Foundry Space.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"space_id", "space_name", "organization_id"},
		)
		spaceTotalAppTasksQuotaMetric.WithLabelValues(spaceId1, spaceName1, organizationId1).Set(float64(AppTaskLimit1))
		spaceTotalAppTasksQuotaMetric.WithLabelValues(spaceId2, spaceName2, organizationId2).Set(float64(AppTaskLimit2))

		spaceTotalMemoryMbQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "space",
				Name:        "total_memory_mb_quota",
				Help:        "Total amount of memory (Mb) a Cloud Foundry Space can have.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"space_id", "space_name", "organization_id"},
		)
		spaceTotalMemoryMbQuotaMetric.WithLabelValues(spaceId1, spaceName1, organizationId1).Set(float64(MemoryLimit1))
		spaceTotalMemoryMbQuotaMetric.WithLabelValues(spaceId2, spaceName2, organizationId2).Set(float64(MemoryLimit2))

		spaceTotalReservedRoutePortsQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "space",
				Name:        "total_reserved_route_ports_quota",
				Help:        "Total number of routes that may be created with reserved ports in a Cloud Foundry Space.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"space_id", "space_name", "organization_id"},
		)
		spaceTotalReservedRoutePortsQuotaMetric.WithLabelValues(spaceId1, spaceName1, organizationId1).Set(float64(TotalReservedRoutePorts1))
		spaceTotalReservedRoutePortsQuotaMetric.WithLabelValues(spaceId2, spaceName2, organizationId2).Set(float64(TotalReservedRoutePorts2))

		spaceTotalRoutesQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "space",
				Name:        "total_routes_quota",
				Help:        "Total number of routes that may be created in a Cloud Foundry Space.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"space_id", "space_name", "organization_id"},
		)
		spaceTotalRoutesQuotaMetric.WithLabelValues(spaceId1, spaceName1, organizationId1).Set(float64(TotalRoutes1))
		spaceTotalRoutesQuotaMetric.WithLabelValues(spaceId2, spaceName2, organizationId2).Set(float64(TotalRoutes2))

		spaceTotalServiceKeysQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "space",
				Name:        "total_service_keys_quota",
				Help:        "Total number of service keys that may be created in a Cloud Foundry Space.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"space_id", "space_name", "organization_id"},
		)
		spaceTotalServiceKeysQuotaMetric.WithLabelValues(spaceId1, spaceName1, organizationId1).Set(float64(TotalServiceKeys1))
		spaceTotalServiceKeysQuotaMetric.WithLabelValues(spaceId2, spaceName2, organizationId2).Set(float64(TotalServiceKeys2))

		spaceTotalServicesQuotaMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "space",
				Name:        "total_services_quota",
				Help:        "Total number of service instances that may be created in a Cloud Foundry Space.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"space_id", "space_name", "organization_id"},
		)
		spaceTotalServicesQuotaMetric.WithLabelValues(spaceId1, spaceName1, organizationId1).Set(float64(TotalServices1))
		spaceTotalServicesQuotaMetric.WithLabelValues(spaceId2, spaceName2, organizationId2).Set(float64(TotalServices2))

		spacesScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "spaces_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Spaces.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		spacesScrapesTotalMetric.Inc()

		spacesScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "spaces_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrapes errors of Cloud Foundry Spaces.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastSpacesScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_spaces_scrape_error",
				Help:        "Whether the last scrape of Spaces metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastSpacesScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_spaces_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Spaces metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastSpacesScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_spaces_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Spaces metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		spacesCollector = NewSpacesCollector(namespace, environment, deployment, cfClient)
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
			go spacesCollector.Describe(descriptions)
		})

		It("returns a space_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spaceInfoMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
				quotaDefinitionName1,
			).Desc())))
		})

		It("returns a space_non_basic_services_allowed metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spaceNonBasicServicesAllowedMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			).Desc())))
		})

		It("returns a space_instance_memory_mb_limit metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spaceInstanceMemoryMbLimitMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			).Desc())))
		})

		It("returns a space_total_app_instances_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spaceTotalAppInstancesQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			).Desc())))
		})

		It("returns a space_total_app_tasks_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spaceTotalAppTasksQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			).Desc())))
		})

		It("returns a space_total_memory_mb_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spaceTotalMemoryMbQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			).Desc())))
		})

		It("returns a space_total_reserved_route_ports_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spaceTotalReservedRoutePortsQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			).Desc())))
		})

		It("returns a space_total_routes_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spaceTotalRoutesQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			).Desc())))
		})

		It("returns a space_total_service_keys_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spaceTotalServiceKeysQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			).Desc())))
		})

		It("returns a space_total_services_quota metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spaceTotalServicesQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			).Desc())))
		})

		It("returns a spaces_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spacesScrapesTotalMetric.Desc())))
		})

		It("returns a spaces_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(spacesScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_spaces_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastSpacesScrapeErrorMetric.Desc())))
		})

		It("returns a last_spaces_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastSpacesScrapeTimestampMetric.Desc())))
		})

		It("returns a last_spaces_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastSpacesScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode          int
			spaceQuotasResponse cfclient.SpaceQuotasResponse
			spacesResponse      cfclient.SpaceResponse
			metrics             chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			spaceQuotasResponse = cfclient.SpaceQuotasResponse{
				Resources: []cfclient.SpaceQuotasResource{
					cfclient.SpaceQuotasResource{
						Meta: cfclient.Meta{
							Guid: quotaDefinitionId1,
						},
						Entity: cfclient.SpaceQuota{
							Name:                    quotaDefinitionName1,
							OrganizationGuid:        organizationId1,
							NonBasicServicesAllowed: false,
							TotalServices:           TotalServices1,
							TotalRoutes:             TotalRoutes1,
							MemoryLimit:             MemoryLimit1,
							InstanceMemoryLimit:     InstanceMemoryLimit1,
							AppInstanceLimit:        AppInstanceLimit1,
							AppTaskLimit:            AppTaskLimit1,
							TotalServiceKeys:        TotalServiceKeys1,
							TotalReservedRoutePorts: TotalReservedRoutePorts1,
						},
					},
					cfclient.SpaceQuotasResource{
						Meta: cfclient.Meta{
							Guid: quotaDefinitionId2,
						},
						Entity: cfclient.SpaceQuota{
							Name:                    quotaDefinitionName2,
							OrganizationGuid:        organizationId2,
							NonBasicServicesAllowed: true,
							TotalServices:           TotalServices2,
							TotalRoutes:             TotalRoutes2,
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

			spacesResponse = cfclient.SpaceResponse{
				Resources: []cfclient.SpaceResource{
					cfclient.SpaceResource{
						Meta: cfclient.Meta{
							Guid: spaceId1,
						},
						Entity: cfclient.Space{
							Name:                spaceName1,
							OrganizationGuid:    organizationId1,
							QuotaDefinitionGuid: quotaDefinitionId1,
						},
					},
					cfclient.SpaceResource{
						Meta: cfclient.Meta{
							Guid: spaceId2,
						},
						Entity: cfclient.Space{
							Name:                spaceName2,
							OrganizationGuid:    organizationId2,
							QuotaDefinitionGuid: quotaDefinitionId2,
						},
					},
					cfclient.SpaceResource{
						Meta: cfclient.Meta{
							Guid: spaceId3,
						},
						Entity: cfclient.Space{
							Name:                spaceName3,
							OrganizationGuid:    organizationId3,
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
					ghttp.VerifyRequest("GET", "/v2/space_quota_definitions"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &spaceQuotasResponse),
				),
				ghttp.CombineHandlers(

					ghttp.VerifyRequest("GET", "/v2/spaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &spacesResponse),
				),
			)

			go spacesCollector.Collect(metrics)
		})

		It("returns a space_info metric for space 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceInfoMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
				quotaDefinitionName1,
			))))
		})

		It("returns a space_info metric for space 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceInfoMetric.WithLabelValues(
				spaceId2,
				spaceName2,
				organizationId2,
				quotaDefinitionName2,
			))))
		})

		It("returns a space_info metric for space 3", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceInfoMetric.WithLabelValues(
				spaceId3,
				spaceName3,
				organizationId3,
				quotaDefinitionName3,
			))))
		})

		It("returns a space_non_basic_services_allowed metric for space 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceNonBasicServicesAllowedMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			))))
		})

		It("returns a space_non_basic_services_allowed metric for space 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceNonBasicServicesAllowedMetric.WithLabelValues(
				spaceId2,
				spaceName2,
				organizationId2,
			))))
		})

		It("does not returns a space_non_basic_services_allowed metric for space 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(spaceNonBasicServicesAllowedMetric.WithLabelValues(
				spaceId3,
				spaceName3,
				organizationId3,
			))))
		})

		It("returns a space_instance_memory_mb_limit metric for space 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceInstanceMemoryMbLimitMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			))))
		})

		It("returns a space_instance_memory_mb_limit metric for space 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceInstanceMemoryMbLimitMetric.WithLabelValues(
				spaceId2,
				spaceName2,
				organizationId2,
			))))
		})

		It("does not returns a space_instance_memory_mb_limit metric for space 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(spaceInstanceMemoryMbLimitMetric.WithLabelValues(
				spaceId3,
				spaceName3,
				organizationId3,
			))))
		})

		It("returns a space_total_app_instances_quota metric for space 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalAppInstancesQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			))))
		})

		It("returns a space_total_app_instances_quota metric for space 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalAppInstancesQuotaMetric.WithLabelValues(
				spaceId2,
				spaceName2,
				organizationId2,
			))))
		})

		It("does not returns a space_total_app_instances_quota metric for space 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(spaceTotalAppInstancesQuotaMetric.WithLabelValues(
				spaceId3,
				spaceName3,
				organizationId3,
			))))
		})

		It("returns a space_total_app_tasks_quota metric for space 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalAppTasksQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			))))
		})

		It("returns a space_total_app_tasks_quota metric for space 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalAppTasksQuotaMetric.WithLabelValues(
				spaceId2,
				spaceName2,
				organizationId2,
			))))
		})

		It("does not returns a space_total_app_tasks_quota metric for space 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(spaceTotalAppTasksQuotaMetric.WithLabelValues(
				spaceId3,
				spaceName3,
				organizationId3,
			))))
		})

		It("returns a space_total_memory_mb_quota metric for space 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalMemoryMbQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			))))
		})

		It("returns a space_total_memory_mb_quota metric for space 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalMemoryMbQuotaMetric.WithLabelValues(
				spaceId2,
				spaceName2,
				organizationId2,
			))))
		})

		It("does not returns a space_total_memory_mb_quota metric for space 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(spaceTotalMemoryMbQuotaMetric.WithLabelValues(
				spaceId3,
				spaceName3,
				organizationId3,
			))))
		})

		It("returns a space_total_reserved_route_ports_quota metric for space 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalReservedRoutePortsQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			))))
		})

		It("returns a space_total_reserved_route_ports_quota metric for space 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalReservedRoutePortsQuotaMetric.WithLabelValues(
				spaceId2,
				spaceName2,
				organizationId2,
			))))
		})

		It("does not returns a space_total_reserved_route_ports_quota metric for space 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(spaceTotalReservedRoutePortsQuotaMetric.WithLabelValues(
				spaceId3,
				spaceName3,
				organizationId3,
			))))
		})

		It("returns a space_total_routes_quota metric for space 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalRoutesQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			))))
		})

		It("returns a space_total_routes_quota metric for space 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalRoutesQuotaMetric.WithLabelValues(
				spaceId2,
				spaceName2,
				organizationId2,
			))))
		})

		It("does not returns a space_total_routes_quota metric for space 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(spaceTotalRoutesQuotaMetric.WithLabelValues(
				spaceId3,
				spaceName3,
				organizationId3,
			))))
		})

		It("returns a space_total_service_keys_quota metric for space 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalServiceKeysQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			))))
		})

		It("returns a space_total_service_keys_quota metric for space 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalServiceKeysQuotaMetric.WithLabelValues(
				spaceId2,
				spaceName2,
				organizationId2,
			))))
		})

		It("does not returns a space_total_service_keys_quota metric for space 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(spaceTotalServiceKeysQuotaMetric.WithLabelValues(
				spaceId3,
				spaceName3,
				organizationId3,
			))))
		})

		It("returns a space_total_services_quota metric for space 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalServicesQuotaMetric.WithLabelValues(
				spaceId1,
				spaceName1,
				organizationId1,
			))))
		})

		It("returns a space_total_services_quota metric for space 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spaceTotalServicesQuotaMetric.WithLabelValues(
				spaceId2,
				spaceName2,
				organizationId2,
			))))
		})

		It("does not returns a space_total_services_quota metric for space 3", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(spaceTotalServicesQuotaMetric.WithLabelValues(
				spaceId3,
				spaceName3,
				organizationId3,
			))))
		})

		It("returns a spaces_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spacesScrapesTotalMetric)))
		})

		It("returns a spaces_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(spacesScrapeErrorsTotalMetric)))
		})

		It("returns a last_spaces_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastSpacesScrapeErrorMetric)))
		})

		Context("when it fails to list the spaces", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				spacesScrapeErrorsTotalMetric.Inc()
				lastSpacesScrapeErrorMetric.Set(1)
			})

			It("returns a spaces_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(spacesScrapeErrorsTotalMetric)))
			})

			It("returns a last_spaces_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastSpacesScrapeErrorMetric)))
			})
		})
	})
})
