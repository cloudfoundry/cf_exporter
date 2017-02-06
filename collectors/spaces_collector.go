package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type SpacesCollector struct {
	namespace                               string
	deploymentName                          string
	cfClient                                *cfclient.Client
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
	spacesTotalMetric                       *prometheus.GaugeVec
	spacesScrapesTotalMetric                *prometheus.CounterVec
	spacesScrapeErrorsTotalMetric           *prometheus.CounterVec
	lastSpacesScrapeErrorMetric             *prometheus.GaugeVec
	lastSpacesScrapeTimestampMetric         *prometheus.GaugeVec
	lastSpacesScrapeDurationSecondsMetric   *prometheus.GaugeVec
}

func NewSpacesCollector(namespace string, deploymentName string, cfClient *cfclient.Client) *SpacesCollector {
	spaceInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Space information with a constant '1' value.",
		},
		[]string{"deployment", "space_id", "space_name"},
	)

	spaceNonBasicServicesAllowedMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "non_basic_services_allowed",
			Help:      "A Cloud Foundry Space can provision instances of paid service plans? (1 for true, 0 for false).",
		},
		[]string{"deployment", "space_id", "space_name"},
	)

	spaceInstanceMemoryMbLimitMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "instance_memory_mb_limit",
			Help:      "Maximum amount of memory (Mb) an application instance can have in a Cloud Foundry Space.",
		},
		[]string{"deployment", "space_id", "space_name"},
	)

	spaceTotalAppInstancesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "total_app_instances_quota",
			Help:      "Total number of application instances that may be created in a Cloud Foundry Space.",
		},
		[]string{"deployment", "space_id", "space_name"},
	)

	spaceTotalAppTasksQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "total_app_tasks_quota",
			Help:      "Total number of application tasks that may be created in a Cloud Foundry Space.",
		},
		[]string{"deployment", "space_id", "space_name"},
	)

	spaceTotalMemoryMbQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "total_memory_mb_quota",
			Help:      "Total amount of memory (Mb) a Cloud Foundry Space can have.",
		},
		[]string{"deployment", "space_id", "space_name"},
	)

	spaceTotalReservedRoutePortsQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "total_reserved_route_ports_quota",
			Help:      "Total number of routes that may be created with reserved ports in a Cloud Foundry Space.",
		},
		[]string{"deployment", "space_id", "space_name"},
	)

	spaceTotalRoutesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "total_routes_quota",
			Help:      "Total number of routes that may be created in a Cloud Foundry Space.",
		},
		[]string{"deployment", "space_id", "space_name"},
	)

	spaceTotalServiceKeysQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "total_service_keys_quota",
			Help:      "Total number of service keys that may be created in a Cloud Foundry Space.",
		},
		[]string{"deployment", "space_id", "space_name"},
	)

	spaceTotalServicesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "total_services_quota",
			Help:      "Total number of service instances that may be created in a Cloud Foundry Space.",
		},
		[]string{"deployment", "space_id", "space_name"},
	)

	spacesTotalMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "spaces",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Spaces.",
		},
		[]string{"deployment"},
	)

	spacesScrapesTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "spaces_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Cloud Foundry Spaces.",
		},
		[]string{"deployment"},
	)

	spacesScrapeErrorsTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "spaces_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrapes errors of Cloud Foundry Spaces.",
		},
		[]string{"deployment"},
	)

	lastSpacesScrapeErrorMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_spaces_scrape_error",
			Help:      "Whether the last scrape of Spaces metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
		[]string{"deployment"},
	)

	lastSpacesScrapeTimestampMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_spaces_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Spaces metrics from Cloud Foundry.",
		},
		[]string{"deployment"},
	)

	lastSpacesScrapeDurationSecondsMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_spaces_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Spaces metrics from Cloud Foundry.",
		},
		[]string{"deployment"},
	)

	return &SpacesCollector{
		namespace:                               namespace,
		deploymentName:                          deploymentName,
		cfClient:                                cfClient,
		spaceInfoMetric:                         spaceInfoMetric,
		spaceNonBasicServicesAllowedMetric:      spaceNonBasicServicesAllowedMetric,
		spaceInstanceMemoryMbLimitMetric:        spaceInstanceMemoryMbLimitMetric,
		spaceTotalAppInstancesQuotaMetric:       spaceTotalAppInstancesQuotaMetric,
		spaceTotalAppTasksQuotaMetric:           spaceTotalAppTasksQuotaMetric,
		spaceTotalMemoryMbQuotaMetric:           spaceTotalMemoryMbQuotaMetric,
		spaceTotalReservedRoutePortsQuotaMetric: spaceTotalReservedRoutePortsQuotaMetric,
		spaceTotalRoutesQuotaMetric:             spaceTotalRoutesQuotaMetric,
		spaceTotalServiceKeysQuotaMetric:        spaceTotalServiceKeysQuotaMetric,
		spaceTotalServicesQuotaMetric:           spaceTotalServicesQuotaMetric,
		spacesTotalMetric:                       spacesTotalMetric,
		spacesScrapesTotalMetric:                spacesScrapesTotalMetric,
		spacesScrapeErrorsTotalMetric:           spacesScrapeErrorsTotalMetric,
		lastSpacesScrapeErrorMetric:             lastSpacesScrapeErrorMetric,
		lastSpacesScrapeTimestampMetric:         lastSpacesScrapeTimestampMetric,
		lastSpacesScrapeDurationSecondsMetric:   lastSpacesScrapeDurationSecondsMetric,
	}
}

func (c SpacesCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportSpacesMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.spacesScrapeErrorsTotalMetric.WithLabelValues(c.deploymentName).Inc()
	}
	c.spacesScrapeErrorsTotalMetric.Collect(ch)

	c.spacesScrapesTotalMetric.WithLabelValues(c.deploymentName).Inc()
	c.spacesScrapesTotalMetric.Collect(ch)

	c.lastSpacesScrapeErrorMetric.WithLabelValues(c.deploymentName).Set(errorMetric)
	c.lastSpacesScrapeErrorMetric.Collect(ch)

	c.lastSpacesScrapeTimestampMetric.WithLabelValues(c.deploymentName).Set(float64(time.Now().Unix()))
	c.lastSpacesScrapeTimestampMetric.Collect(ch)

	c.lastSpacesScrapeDurationSecondsMetric.WithLabelValues(c.deploymentName).Set(time.Since(begun).Seconds())
	c.lastSpacesScrapeDurationSecondsMetric.Collect(ch)
}

func (c SpacesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.spaceInfoMetric.Describe(ch)
	c.spaceNonBasicServicesAllowedMetric.Describe(ch)
	c.spaceInstanceMemoryMbLimitMetric.Describe(ch)
	c.spaceTotalAppInstancesQuotaMetric.Describe(ch)
	c.spaceTotalAppTasksQuotaMetric.Describe(ch)
	c.spaceTotalMemoryMbQuotaMetric.Describe(ch)
	c.spaceTotalReservedRoutePortsQuotaMetric.Describe(ch)
	c.spaceTotalRoutesQuotaMetric.Describe(ch)
	c.spaceTotalServiceKeysQuotaMetric.Describe(ch)
	c.spaceTotalServicesQuotaMetric.Describe(ch)
	c.spacesScrapesTotalMetric.Describe(ch)
	c.spacesScrapeErrorsTotalMetric.Describe(ch)
	c.lastSpacesScrapeErrorMetric.Describe(ch)
	c.lastSpacesScrapeTimestampMetric.Describe(ch)
	c.lastSpacesScrapeDurationSecondsMetric.Describe(ch)
}

func (c SpacesCollector) reportSpacesMetrics(ch chan<- prometheus.Metric) error {
	c.spaceInfoMetric.Reset()
	c.spaceNonBasicServicesAllowedMetric.Reset()
	c.spaceInstanceMemoryMbLimitMetric.Reset()
	c.spaceTotalAppInstancesQuotaMetric.Reset()
	c.spaceTotalAppTasksQuotaMetric.Reset()
	c.spaceTotalMemoryMbQuotaMetric.Reset()
	c.spaceTotalReservedRoutePortsQuotaMetric.Reset()
	c.spaceTotalRoutesQuotaMetric.Reset()
	c.spaceTotalServiceKeysQuotaMetric.Reset()
	c.spaceTotalServicesQuotaMetric.Reset()

	spaceQuotas, err := c.gatherSpaceQuotas()
	if err != nil {
		log.Errorf("Error while listing space quotas: %v", err)
		return err
	}

	spaces, err := c.cfClient.ListSpaces()
	if err != nil {
		log.Errorf("Error while listing spaces: %v", err)
		return err
	}

	for _, space := range spaces {
		c.spaceInfoMetric.WithLabelValues(
			c.deploymentName,
			space.Guid,
			space.Name,
		).Set(float64(1))

		if space.QuotaDefinitionGuid != "" {
			if spaceQuota, ok := spaceQuotas[space.QuotaDefinitionGuid]; ok {
				c.reportSpaceQuotasMetrics(space.Guid, space.Name, spaceQuota)
			}
		}
	}

	c.spaceInfoMetric.Collect(ch)

	c.spaceNonBasicServicesAllowedMetric.Collect(ch)
	c.spaceInstanceMemoryMbLimitMetric.Collect(ch)
	c.spaceTotalAppInstancesQuotaMetric.Collect(ch)
	c.spaceTotalAppTasksQuotaMetric.Collect(ch)
	c.spaceTotalMemoryMbQuotaMetric.Collect(ch)
	c.spaceTotalReservedRoutePortsQuotaMetric.Collect(ch)
	c.spaceTotalRoutesQuotaMetric.Collect(ch)
	c.spaceTotalServiceKeysQuotaMetric.Collect(ch)
	c.spaceTotalServicesQuotaMetric.Collect(ch)

	c.spacesTotalMetric.WithLabelValues(c.deploymentName).Set(float64(len(spaces)))
	c.spacesTotalMetric.Collect(ch)

	return nil
}

func (c SpacesCollector) gatherSpaceQuotas() (map[string]*cfclient.SpaceQuota, error) {
	quotas, err := c.cfClient.ListSpaceQuotas()
	if err != nil {
		return nil, err
	}

	spaceQuotas := make(map[string]*cfclient.SpaceQuota, len(quotas))
	for _, quota := range quotas {
		spaceQuotas[quota.Guid] = &quota
	}

	return spaceQuotas, nil
}

func (c SpacesCollector) reportSpaceQuotasMetrics(spaceGuid string, spaceName string, spaceQuota *cfclient.SpaceQuota) {
	nonBasicServicesAllowed := 0
	if spaceQuota.NonBasicServicesAllowed {
		nonBasicServicesAllowed = 1
	}
	c.spaceNonBasicServicesAllowedMetric.WithLabelValues(
		c.deploymentName,
		spaceGuid,
		spaceName,
	).Set(float64(nonBasicServicesAllowed))

	c.spaceInstanceMemoryMbLimitMetric.WithLabelValues(
		c.deploymentName,
		spaceGuid,
		spaceName,
	).Set(float64(spaceQuota.InstanceMemoryLimit))

	c.spaceTotalAppInstancesQuotaMetric.WithLabelValues(
		c.deploymentName,
		spaceGuid,
		spaceName,
	).Set(float64(spaceQuota.AppInstanceLimit))

	c.spaceTotalAppTasksQuotaMetric.WithLabelValues(
		c.deploymentName,
		spaceGuid,
		spaceName,
	).Set(float64(spaceQuota.AppTaskLimit))

	c.spaceTotalMemoryMbQuotaMetric.WithLabelValues(
		c.deploymentName,
		spaceGuid,
		spaceName,
	).Set(float64(spaceQuota.MemoryLimit))

	c.spaceTotalReservedRoutePortsQuotaMetric.WithLabelValues(
		c.deploymentName,
		spaceGuid,
		spaceName,
	).Set(float64(spaceQuota.TotalReservedRoutePorts))

	c.spaceTotalRoutesQuotaMetric.WithLabelValues(
		c.deploymentName,
		spaceGuid,
		spaceName,
	).Set(float64(spaceQuota.TotalRoutes))

	c.spaceTotalServiceKeysQuotaMetric.WithLabelValues(
		c.deploymentName,
		spaceGuid,
		spaceName,
	).Set(float64(spaceQuota.TotalServiceKeys))

	c.spaceTotalServicesQuotaMetric.WithLabelValues(
		c.deploymentName,
		spaceGuid,
		spaceName,
	).Set(float64(spaceQuota.TotalServices))
}
