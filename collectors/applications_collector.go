package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ApplicationsCollector struct {
	namespace                                   string
	deploymentName                              string
	cfClient                                    *cfclient.Client
	applicationInfoMetric                       *prometheus.GaugeVec
	applicationInstancesMetric                  *prometheus.GaugeVec
	applicationMemoryMbMetric                   *prometheus.GaugeVec
	applicationDiskQuotaMbMetric                *prometheus.GaugeVec
	applicationsScrapesTotalMetric              *prometheus.CounterVec
	applicationsScrapeErrorsTotalMetric         *prometheus.CounterVec
	lastApplicationsScrapeErrorMetric           *prometheus.GaugeVec
	lastApplicationsScrapeTimestampMetric       *prometheus.GaugeVec
	lastApplicationsScrapeDurationSecondsMetric *prometheus.GaugeVec
}

func NewApplicationsCollector(namespace string, deploymentName string, cfClient *cfclient.Client) *ApplicationsCollector {
	applicationInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "application",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Application information with a constant '1' value.",
		},
		[]string{"deployment", "application_id", "application_name", "buildpack", "organization_id", "organization_name", "space_id", "space_name", "stack_id", "state"},
	)

	applicationInstancesMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "application",
			Name:      "instances",
			Help:      "Cloud Foundry Application Instances.",
		},
		[]string{"deployment", "application_id", "application_name", "organization_id", "organization_name", "space_id", "space_name"},
	)

	applicationMemoryMbMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "application",
			Name:      "memory_mb",
			Help:      "Cloud Foundry Application Memory (Mb).",
		},
		[]string{"deployment", "application_id", "application_name", "organization_id", "organization_name", "space_id", "space_name"},
	)

	applicationDiskQuotaMbMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "application",
			Name:      "disk_quota_mb",
			Help:      "Cloud Foundry Application Disk Quota (Mb).",
		},
		[]string{"deployment", "application_id", "application_name", "organization_id", "organization_name", "space_id", "space_name"},
	)

	applicationsScrapesTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "applications_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Cloud Foundry Applications.",
		},
		[]string{"deployment"},
	)

	applicationsScrapeErrorsTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "applications_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrape errors of Cloud Foundry Applications.",
		},
		[]string{"deployment"},
	)

	lastApplicationsScrapeErrorMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_applications_scrape_error",
			Help:      "Whether the last scrape of Applications metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
		[]string{"deployment"},
	)

	lastApplicationsScrapeTimestampMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_applications_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Applications metrics from Cloud Foundry.",
		},
		[]string{"deployment"},
	)

	lastApplicationsScrapeDurationSecondsMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_applications_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Applications metrics from Cloud Foundry.",
		},
		[]string{"deployment"},
	)

	return &ApplicationsCollector{
		namespace:                                   namespace,
		deploymentName:                              deploymentName,
		cfClient:                                    cfClient,
		applicationInfoMetric:                       applicationInfoMetric,
		applicationInstancesMetric:                  applicationInstancesMetric,
		applicationMemoryMbMetric:                   applicationMemoryMbMetric,
		applicationDiskQuotaMbMetric:                applicationDiskQuotaMbMetric,
		applicationsScrapesTotalMetric:              applicationsScrapesTotalMetric,
		applicationsScrapeErrorsTotalMetric:         applicationsScrapeErrorsTotalMetric,
		lastApplicationsScrapeErrorMetric:           lastApplicationsScrapeErrorMetric,
		lastApplicationsScrapeTimestampMetric:       lastApplicationsScrapeTimestampMetric,
		lastApplicationsScrapeDurationSecondsMetric: lastApplicationsScrapeDurationSecondsMetric,
	}
}

func (c ApplicationsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportApplicationsMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.applicationsScrapeErrorsTotalMetric.WithLabelValues(c.deploymentName).Inc()
	}
	c.applicationsScrapeErrorsTotalMetric.Collect(ch)

	c.applicationsScrapesTotalMetric.WithLabelValues(c.deploymentName).Inc()
	c.applicationsScrapesTotalMetric.Collect(ch)

	c.lastApplicationsScrapeErrorMetric.WithLabelValues(c.deploymentName).Set(errorMetric)
	c.lastApplicationsScrapeErrorMetric.Collect(ch)

	c.lastApplicationsScrapeTimestampMetric.WithLabelValues(c.deploymentName).Set(float64(time.Now().Unix()))
	c.lastApplicationsScrapeTimestampMetric.Collect(ch)

	c.lastApplicationsScrapeDurationSecondsMetric.WithLabelValues(c.deploymentName).Set(time.Since(begun).Seconds())
	c.lastApplicationsScrapeDurationSecondsMetric.Collect(ch)
}

func (c ApplicationsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.applicationInfoMetric.Describe(ch)
	c.applicationInstancesMetric.Describe(ch)
	c.applicationMemoryMbMetric.Describe(ch)
	c.applicationDiskQuotaMbMetric.Describe(ch)
	c.applicationsScrapesTotalMetric.Describe(ch)
	c.applicationsScrapeErrorsTotalMetric.Describe(ch)
	c.lastApplicationsScrapeErrorMetric.Describe(ch)
	c.lastApplicationsScrapeTimestampMetric.Describe(ch)
	c.lastApplicationsScrapeDurationSecondsMetric.Describe(ch)
}

func (c ApplicationsCollector) reportApplicationsMetrics(ch chan<- prometheus.Metric) error {
	c.applicationInfoMetric.Reset()
	c.applicationInstancesMetric.Reset()
	c.applicationMemoryMbMetric.Reset()
	c.applicationDiskQuotaMbMetric.Reset()

	applications, err := c.cfClient.ListApps()
	if err != nil {
		log.Errorf("Error while listing applications: %v", err)
		return err
	}

	for _, application := range applications {
		c.applicationInfoMetric.WithLabelValues(
			c.deploymentName,
			application.Guid,
			application.Name,
			application.Buildpack,
			application.SpaceData.Entity.OrgData.Entity.Guid,
			application.SpaceData.Entity.OrgData.Entity.Name,
			application.SpaceData.Entity.Guid,
			application.SpaceData.Entity.Name,
			application.StackGuid,
			application.State,
		).Set(float64(1))

		c.applicationMemoryMbMetric.WithLabelValues(
			c.deploymentName,
			application.Guid,
			application.Name,
			application.SpaceData.Entity.OrgData.Entity.Guid,
			application.SpaceData.Entity.OrgData.Entity.Name,
			application.SpaceData.Entity.Guid,
			application.SpaceData.Entity.Name,
		).Set(float64(application.Memory))

		c.applicationInstancesMetric.WithLabelValues(
			c.deploymentName,
			application.Guid,
			application.Name,
			application.SpaceData.Entity.OrgData.Entity.Guid,
			application.SpaceData.Entity.OrgData.Entity.Name,
			application.SpaceData.Entity.Guid,
			application.SpaceData.Entity.Name,
		).Set(float64(application.Instances))

		c.applicationDiskQuotaMbMetric.WithLabelValues(
			c.deploymentName,
			application.Guid,
			application.Name,
			application.SpaceData.Entity.OrgData.Entity.Guid,
			application.SpaceData.Entity.OrgData.Entity.Name,
			application.SpaceData.Entity.Guid,
			application.SpaceData.Entity.Name,
		).Set(float64(application.DiskQuota))
	}

	c.applicationInfoMetric.Collect(ch)
	c.applicationInstancesMetric.Collect(ch)
	c.applicationMemoryMbMetric.Collect(ch)
	c.applicationDiskQuotaMbMetric.Collect(ch)

	return nil
}
