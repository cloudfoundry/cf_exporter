package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ApplicationsCollector struct {
	namespace                                   string
	cfClient                                    *cfclient.Client
	applicationInfoMetric                       *prometheus.GaugeVec
	applicationsTotalMetric                     prometheus.Gauge
	lastApplicationsScrapeErrorMetric           prometheus.Gauge
	lastApplicationsScrapeTimestampMetric       prometheus.Gauge
	lastApplicationsScrapeDurationSecondsMetric prometheus.Gauge
}

func NewApplicationsCollector(namespace string, cfClient *cfclient.Client) *ApplicationsCollector {
	applicationInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "application",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Application information with a constant '1' value.",
		},
		[]string{"application_id", "application_name", "space_id", "space_name", "organization_id", "organization_name"},
	)

	applicationsTotalMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "applications",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Applications.",
		},
	)

	lastApplicationsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_applications_scrape_error",
			Help:      "Whether the last scrape of Applications metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastApplicationsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_applications_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Applications metrics from Cloud Foundry.",
		},
	)

	lastApplicationsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_applications_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Applications metrics from Cloud Foundry.",
		},
	)

	return &ApplicationsCollector{
		namespace:                                   namespace,
		cfClient:                                    cfClient,
		applicationInfoMetric:                       applicationInfoMetric,
		applicationsTotalMetric:                     applicationsTotalMetric,
		lastApplicationsScrapeErrorMetric:           lastApplicationsScrapeErrorMetric,
		lastApplicationsScrapeTimestampMetric:       lastApplicationsScrapeTimestampMetric,
		lastApplicationsScrapeDurationSecondsMetric: lastApplicationsScrapeDurationSecondsMetric,
	}
}

func (c ApplicationsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	c.applicationInfoMetric.Reset()

	applications, err := c.cfClient.ListApps()
	if err != nil {
		log.Errorf("Error while listing applications: %v", err)
		c.reportErrorMetric(true, ch)
		return
	}

	for _, application := range applications {
		c.applicationInfoMetric.WithLabelValues(
			application.Guid,
			application.Name,
			application.SpaceData.Entity.Guid,
			application.SpaceData.Entity.Name,
			application.SpaceData.Entity.OrgData.Entity.Guid,
			application.SpaceData.Entity.OrgData.Entity.Name,
		).Set(float64(1))
	}

	c.applicationInfoMetric.Collect(ch)

	c.applicationsTotalMetric.Set(float64(len(applications)))
	c.applicationsTotalMetric.Collect(ch)

	c.lastApplicationsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastApplicationsScrapeTimestampMetric.Collect(ch)

	c.lastApplicationsScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastApplicationsScrapeDurationSecondsMetric.Collect(ch)

	c.reportErrorMetric(false, ch)
}

func (c ApplicationsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.applicationInfoMetric.Describe(ch)
	c.applicationsTotalMetric.Describe(ch)
	c.lastApplicationsScrapeErrorMetric.Describe(ch)
	c.lastApplicationsScrapeTimestampMetric.Describe(ch)
	c.lastApplicationsScrapeDurationSecondsMetric.Describe(ch)
}

func (c ApplicationsCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if errorHappend {
		errorMetric = float64(1)
	}

	c.lastApplicationsScrapeErrorMetric.Set(errorMetric)
	c.lastApplicationsScrapeErrorMetric.Collect(ch)
}
