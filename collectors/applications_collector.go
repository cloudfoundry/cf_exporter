package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ApplicationsCollector struct {
	namespace                             string
	cfClient                              *cfclient.Client
	applicationInfo                       *prometheus.GaugeVec
	applicationsTotal                     prometheus.Gauge
	lastApplicationsScrapeError           prometheus.Gauge
	lastApplicationsScrapeTimestamp       prometheus.Gauge
	lastApplicationsScrapeDurationSeconds prometheus.Gauge
}

func NewApplicationsCollector(namespace string, cfClient *cfclient.Client) *ApplicationsCollector {
	applicationInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "application",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Application information with a constant '1' value.",
		},
		[]string{"application_id", "application_name", "space_id", "space_name", "organization_id", "organization_name"},
	)

	applicationsTotal := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "applications",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Applications.",
		},
	)

	lastApplicationsScrapeError := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_applications_scrape_error",
			Help:      "Whether the last scrape of Applications metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastApplicationsScrapeTimestamp := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_applications_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Applications metrics from Cloud Foundry.",
		},
	)

	lastApplicationsScrapeDurationSeconds := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_applications_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Applications metrics from Cloud Foundry.",
		},
	)

	return &ApplicationsCollector{
		namespace:                             namespace,
		cfClient:                              cfClient,
		applicationInfo:                       applicationInfo,
		applicationsTotal:                     applicationsTotal,
		lastApplicationsScrapeError:           lastApplicationsScrapeError,
		lastApplicationsScrapeTimestamp:       lastApplicationsScrapeTimestamp,
		lastApplicationsScrapeDurationSeconds: lastApplicationsScrapeDurationSeconds,
	}
}

func (c ApplicationsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	applications, err := c.cfClient.ListApps()
	if err != nil {
		log.Errorf("Error while listing applications: %v", err)
		c.reportErrorMetric(true, ch)
		return
	}

	for _, application := range applications {
		c.applicationInfo.WithLabelValues(
			application.Guid,
			application.Name,
			application.SpaceData.Entity.Guid,
			application.SpaceData.Entity.Name,
			application.SpaceData.Entity.OrgData.Entity.Guid,
			application.SpaceData.Entity.OrgData.Entity.Name,
		).Set(float64(1))
	}
	c.applicationInfo.Collect(ch)

	c.applicationsTotal.Set(float64(len(applications)))
	c.applicationsTotal.Collect(ch)

	c.lastApplicationsScrapeTimestamp.Set(float64(time.Now().Unix()))
	c.lastApplicationsScrapeTimestamp.Collect(ch)

	c.lastApplicationsScrapeDurationSeconds.Set(time.Since(begun).Seconds())
	c.lastApplicationsScrapeDurationSeconds.Collect(ch)

	c.reportErrorMetric(false, ch)
}

func (c ApplicationsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.applicationInfo.Describe(ch)
	c.applicationsTotal.Describe(ch)
	c.lastApplicationsScrapeError.Describe(ch)
	c.lastApplicationsScrapeTimestamp.Describe(ch)
	c.lastApplicationsScrapeDurationSeconds.Describe(ch)
}

func (c ApplicationsCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if errorHappend {
		errorMetric = float64(1)
	}

	c.lastApplicationsScrapeError.Set(errorMetric)
	c.lastApplicationsScrapeError.Collect(ch)
}
