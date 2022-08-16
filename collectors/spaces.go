package collectors

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"github.com/bosh-prometheus/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type SpacesCollector struct {
	namespace                               string
	environment                             string
	deployment                              string
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
}

func NewSpacesCollector(
	namespace string,
	environment string,
	deployment string,
) *SpacesCollector {
	spaceInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "space",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Space information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"space_id", "space_name", "organization_id", "quota_name"},
	)

	spaceNonBasicServicesAllowedMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "space",
			Name:        "non_basic_services_allowed",
			Help:        "A Cloud Foundry Space can provision instances of paid service plans? (1 for true, 0 for false).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"space_id", "space_name", "organization_id"},
	)

	spaceInstanceMemoryMbLimitMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "space",
			Name:        "instance_memory_mb_limit",
			Help:        "Maximum amount of memory (Mb) an application instance can have in a Cloud Foundry Space.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"space_id", "space_name", "organization_id"},
	)

	spaceTotalAppInstancesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "space",
			Name:        "total_app_instances_quota",
			Help:        "Total number of application instances that may be created in a Cloud Foundry Space.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"space_id", "space_name", "organization_id"},
	)

	spaceTotalAppTasksQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "space",
			Name:        "total_app_tasks_quota",
			Help:        "Total number of application tasks that may be created in a Cloud Foundry Space.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"space_id", "space_name", "organization_id"},
	)

	spaceTotalMemoryMbQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "space",
			Name:        "total_memory_mb_quota",
			Help:        "Total amount of memory (Mb) a Cloud Foundry Space can have.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"space_id", "space_name", "organization_id"},
	)

	spaceTotalReservedRoutePortsQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "space",
			Name:        "total_reserved_route_ports_quota",
			Help:        "Total number of routes that may be created with reserved ports in a Cloud Foundry Space.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"space_id", "space_name", "organization_id"},
	)

	spaceTotalRoutesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "space",
			Name:        "total_routes_quota",
			Help:        "Total number of routes that may be created in a Cloud Foundry Space.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"space_id", "space_name", "organization_id"},
	)

	spaceTotalServiceKeysQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "space",
			Name:        "total_service_keys_quota",
			Help:        "Total number of service keys that may be created in a Cloud Foundry Space.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"space_id", "space_name", "organization_id"},
	)

	spaceTotalServicesQuotaMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "space",
			Name:        "total_services_quota",
			Help:        "Total number of service instances that may be created in a Cloud Foundry Space.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"space_id", "space_name", "organization_id"},
	)

	spacesScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "spaces_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Spaces.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	spacesScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "spaces_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrapes errors of Cloud Foundry Spaces.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastSpacesScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_spaces_scrape_error",
			Help:        "Whether the last scrape of Spaces metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastSpacesScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_spaces_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Spaces metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastSpacesScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_spaces_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Spaces metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &SpacesCollector{
		namespace:                               namespace,
		environment:                             environment,
		deployment:                              deployment,
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
		spacesScrapesTotalMetric:                spacesScrapesTotalMetric,
		spacesScrapeErrorsTotalMetric:           spacesScrapeErrorsTotalMetric,
		lastSpacesScrapeErrorMetric:             lastSpacesScrapeErrorMetric,
		lastSpacesScrapeTimestampMetric:         lastSpacesScrapeTimestampMetric,
		lastSpacesScrapeDurationSecondsMetric:   lastSpacesScrapeDurationSecondsMetric,
	}
}

func (c SpacesCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	err := objs.Error
	if objs.Error != nil {
		errorMetric = float64(1)
		c.spacesScrapeErrorsTotalMetric.Inc()
	} else {
		err = c.reportSpacesMetrics(objs, ch)
		if err != nil {
			log.Error(err)
			errorMetric = float64(1)
			c.spacesScrapeErrorsTotalMetric.Inc()
		}
	}

	c.spacesScrapeErrorsTotalMetric.Collect(ch)
	c.spacesScrapesTotalMetric.Inc()
	c.spacesScrapesTotalMetric.Collect(ch)
	c.lastSpacesScrapeErrorMetric.Set(errorMetric)
	c.lastSpacesScrapeErrorMetric.Collect(ch)
	c.lastSpacesScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastSpacesScrapeTimestampMetric.Collect(ch)
	c.lastSpacesScrapeDurationSecondsMetric.Set(objs.Took)
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

// reportSpacesMetrics
// 1. rely on GUID value instead of map status because it
//    may exists in relationship but with empty value
func (c SpacesCollector) reportSpacesMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) error {
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

	for _, cSpace := range objs.Spaces {

		relOrg, ok := cSpace.Relationships[constant.RelationshipTypeOrganization]
		if !ok {
			return fmt.Errorf("could not find org relationship in space '%s'", cSpace.GUID)
		}
		quotaName := ""
		// 1.
		relQuota := cSpace.Relationships[constant.RelationshipTypeQuota]
		if relQuota.GUID != "" {
			quota, okQ := objs.SpaceQuotas[relQuota.GUID]
			if !okQ {
				return fmt.Errorf("could not find space quota '%s' from space '%s'", relQuota.GUID, cSpace.GUID)
			}
			quotaName = quota.Name
			c.spaceNonBasicServicesAllowedMetric.WithLabelValues(
				cSpace.GUID,
				cSpace.Name,
				relOrg.GUID,
			).Set(BoolToFloat(quota.Services.PaidServicePlans))

			c.spaceInstanceMemoryMbLimitMetric.WithLabelValues(
				cSpace.GUID,
				cSpace.Name,
				relOrg.GUID,
			).Set(NullIntToFloat(quota.Apps.InstanceMemory))

			c.spaceTotalAppInstancesQuotaMetric.WithLabelValues(
				cSpace.GUID,
				cSpace.Name,
				relOrg.GUID,
			).Set(NullIntToFloat(quota.Apps.TotalAppInstances))

			c.spaceTotalAppTasksQuotaMetric.WithLabelValues(
				cSpace.GUID,
				cSpace.Name,
				relOrg.GUID,
			).Set(NullIntToFloat(quota.Apps.PerAppTasks))

			c.spaceTotalMemoryMbQuotaMetric.WithLabelValues(
				cSpace.GUID,
				cSpace.Name,
				relOrg.GUID,
			).Set(NullIntToFloat(quota.Apps.TotalMemory))

			c.spaceTotalReservedRoutePortsQuotaMetric.WithLabelValues(
				cSpace.GUID,
				cSpace.Name,
				relOrg.GUID,
			).Set(NullIntToFloat(quota.Routes.TotalReservedPorts))

			c.spaceTotalRoutesQuotaMetric.WithLabelValues(
				cSpace.GUID,
				cSpace.Name,
				relOrg.GUID,
			).Set(NullIntToFloat(quota.Routes.TotalRoutes))

			c.spaceTotalServiceKeysQuotaMetric.WithLabelValues(
				cSpace.GUID,
				cSpace.Name,
				relOrg.GUID,
			).Set(NullIntToFloat(quota.Services.TotalServiceKeys))

			c.spaceTotalServicesQuotaMetric.WithLabelValues(
				cSpace.GUID,
				cSpace.Name,
				relOrg.GUID,
			).Set(NullIntToFloat(quota.Services.TotalServiceInstances))
		}

		c.spaceInfoMetric.WithLabelValues(
			cSpace.GUID,
			cSpace.Name,
			relOrg.GUID,
			quotaName,
		).Set(float64(1))
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
	return nil
}
