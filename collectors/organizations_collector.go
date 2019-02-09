package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type OrganizationsCollector struct {
	namespace                                      string
	environment                                    string
	deployment                                     string
	cfClient                                       *cfclient.Client
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
}

func NewOrganizationsCollector(
	namespace string,
	environment string,
	deployment string,
	cfClient *cfclient.Client,
) *OrganizationsCollector {
	organizationInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Organization information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name", "quota_name"},
	)

	organizationNonBasicServicesAllowedMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "non_basic_services_allowed",
			Help:        "A Cloud Foundry Organization can provision instances of paid service plans? (1 for true, 0 for false).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationInstanceMemoryMbLimitMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "instance_memory_mb_limit",
			Help:        "Maximum amount of memory (Mb) an application instance can have in a Cloud Foundry Organization.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationTotalAppInstancesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "total_app_instances_quota",
			Help:        "Total number of application instances that may be created in a Cloud Foundry Organization.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationTotalAppTasksQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "total_app_tasks_quota",
			Help:        "Total number of application tasks that may be created in a Cloud Foundry Organization.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationTotalMemoryMbQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "total_memory_mb_quota",
			Help:        "Total amount of memory (Mb) a Cloud Foundry Organization can have.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationTotalPrivateDomainsQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "total_private_domains_quota",
			Help:        "Total number of private domains that may be created in a Cloud Foundry Organization.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationTotalReservedRoutePortsQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "total_reserved_route_ports_quota",
			Help:        "Total number of routes that may be created with reserved ports in a Cloud Foundry Organization.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationTotalRoutesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "total_routes_quota",
			Help:        "Total number of routes that may be created in a Cloud Foundry Organization.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationTotalServiceKeysQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "total_service_keys_quota",
			Help:        "Total number of service keys that may be created in a Cloud Foundry Organization.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationTotalServicesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "organization",
			Name:        "total_services_quota",
			Help:        "Total number of service instances that may be created in a Cloud Foundry Organization.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "organizations_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Organizations.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	organizationsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "organizations_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape errors of Cloud Foundry Organizations.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastOrganizationsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_organizations_scrape_error",
			Help:        "Whether the last scrape of Organizations metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastOrganizationsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_organizations_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Organizations metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastOrganizationsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_organizations_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Organizations metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &OrganizationsCollector{
		namespace:                                      namespace,
		environment:                                    environment,
		deployment:                                     deployment,
		cfClient:                                       cfClient,
		organizationInfoMetric:                         organizationInfoMetric,
		organizationNonBasicServicesAllowedMetric:      organizationNonBasicServicesAllowedMetric,
		organizationInstanceMemoryMbLimitMetric:        organizationInstanceMemoryMbLimitMetric,
		organizationTotalAppInstancesQuotaMetric:       organizationTotalAppInstancesQuotaMetric,
		organizationTotalAppTasksQuotaMetric:           organizationTotalAppTasksQuotaMetric,
		organizationTotalMemoryMbQuotaMetric:           organizationTotalMemoryMbQuotaMetric,
		organizationTotalPrivateDomainsQuotaMetric:     organizationTotalPrivateDomainsQuotaMetric,
		organizationTotalReservedRoutePortsQuotaMetric: organizationTotalReservedRoutePortsQuotaMetric,
		organizationTotalRoutesQuotaMetric:             organizationTotalRoutesQuotaMetric,
		organizationTotalServiceKeysQuotaMetric:        organizationTotalServiceKeysQuotaMetric,
		organizationTotalServicesQuotaMetric:           organizationTotalServicesQuotaMetric,
		organizationsScrapesTotalMetric:                organizationsScrapesTotalMetric,
		organizationsScrapeErrorsTotalMetric:           organizationsScrapeErrorsTotalMetric,
		lastOrganizationsScrapeErrorMetric:             lastOrganizationsScrapeErrorMetric,
		lastOrganizationsScrapeTimestampMetric:         lastOrganizationsScrapeTimestampMetric,
		lastOrganizationsScrapeDurationSecondsMetric:   lastOrganizationsScrapeDurationSecondsMetric,
	}
}

func (c OrganizationsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportOrganizationsMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.organizationsScrapeErrorsTotalMetric.Inc()
	}
	c.organizationsScrapeErrorsTotalMetric.Collect(ch)

	c.organizationsScrapesTotalMetric.Inc()
	c.organizationsScrapesTotalMetric.Collect(ch)

	c.lastOrganizationsScrapeErrorMetric.Set(errorMetric)
	c.lastOrganizationsScrapeErrorMetric.Collect(ch)

	c.lastOrganizationsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastOrganizationsScrapeTimestampMetric.Collect(ch)

	c.lastOrganizationsScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastOrganizationsScrapeDurationSecondsMetric.Collect(ch)
}

func (c OrganizationsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.organizationInfoMetric.Describe(ch)
	c.organizationNonBasicServicesAllowedMetric.Describe(ch)
	c.organizationInstanceMemoryMbLimitMetric.Describe(ch)
	c.organizationTotalAppInstancesQuotaMetric.Describe(ch)
	c.organizationTotalAppTasksQuotaMetric.Describe(ch)
	c.organizationTotalMemoryMbQuotaMetric.Describe(ch)
	c.organizationTotalPrivateDomainsQuotaMetric.Describe(ch)
	c.organizationTotalReservedRoutePortsQuotaMetric.Describe(ch)
	c.organizationTotalRoutesQuotaMetric.Describe(ch)
	c.organizationTotalServiceKeysQuotaMetric.Describe(ch)
	c.organizationTotalServicesQuotaMetric.Describe(ch)
	c.organizationsScrapesTotalMetric.Describe(ch)
	c.organizationsScrapeErrorsTotalMetric.Describe(ch)
	c.lastOrganizationsScrapeErrorMetric.Describe(ch)
	c.lastOrganizationsScrapeTimestampMetric.Describe(ch)
	c.lastOrganizationsScrapeDurationSecondsMetric.Describe(ch)
}

func (c OrganizationsCollector) reportOrganizationsMetrics(ch chan<- prometheus.Metric) error {
	c.organizationInfoMetric.Reset()
	c.organizationNonBasicServicesAllowedMetric.Reset()
	c.organizationInstanceMemoryMbLimitMetric.Reset()
	c.organizationTotalAppInstancesQuotaMetric.Reset()
	c.organizationTotalAppTasksQuotaMetric.Reset()
	c.organizationTotalMemoryMbQuotaMetric.Reset()
	c.organizationTotalPrivateDomainsQuotaMetric.Reset()
	c.organizationTotalReservedRoutePortsQuotaMetric.Reset()
	c.organizationTotalRoutesQuotaMetric.Reset()
	c.organizationTotalServiceKeysQuotaMetric.Reset()
	c.organizationTotalServicesQuotaMetric.Reset()

	organizationQuotas, err := c.gatherOrganizationQuotas()
	if err != nil {
		log.Errorf("Error while listing organization quotas: %v", err)
		return err
	}

	organizations, err := c.cfClient.ListOrgs()
	if err != nil {
		log.Errorf("Error while listing organizations: %v", err)
		return err
	}

	for _, organization := range organizations {
		var organizationQuota cfclient.OrgQuota
		var ok bool

		if organization.QuotaDefinitionGuid != "" {
			if organizationQuota, ok = organizationQuotas[organization.QuotaDefinitionGuid]; ok {
				c.reportOrganizationQuotasMetrics(organization.Guid, organization.Name, organizationQuota)

			}
		}
		c.organizationInfoMetric.WithLabelValues(
			organization.Guid,
			organization.Name,
			organizationQuota.Name,
		).Set(float64(1))

	}

	c.organizationInfoMetric.Collect(ch)

	c.organizationNonBasicServicesAllowedMetric.Collect(ch)
	c.organizationInstanceMemoryMbLimitMetric.Collect(ch)
	c.organizationTotalAppInstancesQuotaMetric.Collect(ch)
	c.organizationTotalAppTasksQuotaMetric.Collect(ch)
	c.organizationTotalMemoryMbQuotaMetric.Collect(ch)
	c.organizationTotalPrivateDomainsQuotaMetric.Collect(ch)
	c.organizationTotalReservedRoutePortsQuotaMetric.Collect(ch)
	c.organizationTotalRoutesQuotaMetric.Collect(ch)
	c.organizationTotalServiceKeysQuotaMetric.Collect(ch)
	c.organizationTotalServicesQuotaMetric.Collect(ch)

	return nil
}

func (c OrganizationsCollector) gatherOrganizationQuotas() (map[string]cfclient.OrgQuota, error) {
	quotas, err := c.cfClient.ListOrgQuotas()
	if err != nil {
		return nil, err
	}

	orgQuotas := make(map[string]cfclient.OrgQuota, len(quotas))
	for _, quota := range quotas {
		orgQuotas[quota.Guid] = quota
	}

	return orgQuotas, nil
}

func (c OrganizationsCollector) reportOrganizationQuotasMetrics(orgGuid string, orgName string, orgQuota cfclient.OrgQuota) {
	nonBasicServicesAllowed := 0
	if orgQuota.NonBasicServicesAllowed {
		nonBasicServicesAllowed = 1
	}
	c.organizationNonBasicServicesAllowedMetric.WithLabelValues(
		orgGuid,
		orgName,
	).Set(float64(nonBasicServicesAllowed))

	c.organizationInstanceMemoryMbLimitMetric.WithLabelValues(
		orgGuid,
		orgName,
	).Set(float64(orgQuota.InstanceMemoryLimit))

	c.organizationTotalAppInstancesQuotaMetric.WithLabelValues(
		orgGuid,
		orgName,
	).Set(float64(orgQuota.AppInstanceLimit))

	c.organizationTotalAppTasksQuotaMetric.WithLabelValues(
		orgGuid,
		orgName,
	).Set(float64(orgQuota.AppTaskLimit))

	c.organizationTotalMemoryMbQuotaMetric.WithLabelValues(
		orgGuid,
		orgName,
	).Set(float64(orgQuota.MemoryLimit))

	c.organizationTotalPrivateDomainsQuotaMetric.WithLabelValues(
		orgGuid,
		orgName,
	).Set(float64(orgQuota.TotalPrivateDomains))

	c.organizationTotalReservedRoutePortsQuotaMetric.WithLabelValues(
		orgGuid,
		orgName,
	).Set(float64(orgQuota.TotalReservedRoutePorts))

	c.organizationTotalRoutesQuotaMetric.WithLabelValues(
		orgGuid,
		orgName,
	).Set(float64(orgQuota.TotalRoutes))

	c.organizationTotalServiceKeysQuotaMetric.WithLabelValues(
		orgGuid,
		orgName,
	).Set(float64(orgQuota.TotalServiceKeys))

	c.organizationTotalServicesQuotaMetric.WithLabelValues(
		orgGuid,
		orgName,
	).Set(float64(orgQuota.TotalServices))
}
