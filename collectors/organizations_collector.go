package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type OrganizationsCollector struct {
	namespace                                    string
	cfClient                                     *cfclient.Client
	organizationInfoMetric                       *prometheus.GaugeVec
	organizationsTotalMetric                     prometheus.Gauge
	organizationsScrapesTotalMetric              prometheus.Counter
	organizationsScrapeErrorsTotalMetric         prometheus.Counter
	lastOrganizationsScrapeErrorMetric           prometheus.Gauge
	lastOrganizationsScrapeTimestampMetric       prometheus.Gauge
	lastOrganizationsScrapeDurationSecondsMetric prometheus.Gauge
}

func NewOrganizationsCollector(namespace string, cfClient *cfclient.Client) *OrganizationsCollector {
	organizationInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "organization",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Organization information with a constant '1' value.",
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationsTotalMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "organizations",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Organizations.",
		},
	)

	organizationsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "organizations_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Cloud Foundry Organizations.",
		},
	)

	organizationsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "organizations_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrape errors of Cloud Foundry Organizations.",
		},
	)

	lastOrganizationsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_organizations_scrape_error",
			Help:      "Whether the last scrape of Organizations metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastOrganizationsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_organizations_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Organizations metrics from Cloud Foundry.",
		},
	)

	lastOrganizationsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_organizations_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Organizations metrics from Cloud Foundry.",
		},
	)

	return &OrganizationsCollector{
		namespace:                                    namespace,
		cfClient:                                     cfClient,
		organizationInfoMetric:                       organizationInfoMetric,
		organizationsTotalMetric:                     organizationsTotalMetric,
		organizationsScrapesTotalMetric:              organizationsScrapesTotalMetric,
		organizationsScrapeErrorsTotalMetric:         organizationsScrapeErrorsTotalMetric,
		lastOrganizationsScrapeErrorMetric:           lastOrganizationsScrapeErrorMetric,
		lastOrganizationsScrapeTimestampMetric:       lastOrganizationsScrapeTimestampMetric,
		lastOrganizationsScrapeDurationSecondsMetric: lastOrganizationsScrapeDurationSecondsMetric,
	}
}

func (c OrganizationsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportOrganizationsMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.organizationsScrapeErrorsTotalMetric.Inc()
	}

	c.organizationsScrapesTotalMetric.Inc()
	c.organizationsScrapesTotalMetric.Collect(ch)

	c.organizationsScrapeErrorsTotalMetric.Collect(ch)

	c.lastOrganizationsScrapeErrorMetric.Set(errorMetric)
	c.lastOrganizationsScrapeErrorMetric.Collect(ch)

	c.lastOrganizationsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastOrganizationsScrapeTimestampMetric.Collect(ch)

	c.lastOrganizationsScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastOrganizationsScrapeDurationSecondsMetric.Collect(ch)
}

func (c OrganizationsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.organizationInfoMetric.Describe(ch)
	c.organizationsTotalMetric.Describe(ch)
	c.organizationsScrapesTotalMetric.Describe(ch)
	c.organizationsScrapeErrorsTotalMetric.Describe(ch)
	c.lastOrganizationsScrapeErrorMetric.Describe(ch)
	c.lastOrganizationsScrapeTimestampMetric.Describe(ch)
	c.lastOrganizationsScrapeDurationSecondsMetric.Describe(ch)
}

func (c OrganizationsCollector) reportOrganizationsMetrics(ch chan<- prometheus.Metric) error {
	c.organizationInfoMetric.Reset()

	organizations, err := c.cfClient.ListOrgs()
	if err != nil {
		log.Errorf("Error while listing organizations: %v", err)
		return err
	}

	for _, organization := range organizations {
		c.organizationInfoMetric.WithLabelValues(
			organization.Guid,
			organization.Name,
		).Set(float64(1))
	}

	c.organizationInfoMetric.Collect(ch)

	c.organizationsTotalMetric.Set(float64(len(organizations)))
	c.organizationsTotalMetric.Collect(ch)

	return nil
}
