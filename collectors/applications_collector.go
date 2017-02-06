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
	applicationsTotalMetric                     *prometheus.GaugeVec
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
		[]string{"deployment", "application_id", "application_name", "space_id", "space_name", "organization_id", "organization_name"},
	)

	applicationsTotalMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "applications",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Applications.",
		},
		[]string{"deployment"},
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
		applicationsTotalMetric:                     applicationsTotalMetric,
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
	c.applicationsTotalMetric.Describe(ch)
	c.applicationsScrapesTotalMetric.Describe(ch)
	c.applicationsScrapeErrorsTotalMetric.Describe(ch)
	c.lastApplicationsScrapeErrorMetric.Describe(ch)
	c.lastApplicationsScrapeTimestampMetric.Describe(ch)
	c.lastApplicationsScrapeDurationSecondsMetric.Describe(ch)
}

func (c ApplicationsCollector) reportApplicationsMetrics(ch chan<- prometheus.Metric) error {
	c.applicationInfoMetric.Reset()

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
			application.SpaceData.Entity.Guid,
			application.SpaceData.Entity.Name,
			application.SpaceData.Entity.OrgData.Entity.Guid,
			application.SpaceData.Entity.OrgData.Entity.Name,
		).Set(float64(1))
	}

	c.applicationInfoMetric.Collect(ch)

	c.applicationsTotalMetric.WithLabelValues(c.deploymentName).Set(float64(len(applications)))
	c.applicationsTotalMetric.Collect(ch)

	return nil
}
