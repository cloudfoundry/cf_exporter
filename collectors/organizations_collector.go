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
		lastOrganizationsScrapeErrorMetric:           lastOrganizationsScrapeErrorMetric,
		lastOrganizationsScrapeTimestampMetric:       lastOrganizationsScrapeTimestampMetric,
		lastOrganizationsScrapeDurationSecondsMetric: lastOrganizationsScrapeDurationSecondsMetric,
	}
}

func (c OrganizationsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	c.organizationInfoMetric.Reset()

	organizations, err := c.cfClient.ListOrgs()
	if err != nil {
		log.Errorf("Error while listing organizations: %v", err)
		c.reportErrorMetric(true, ch)
		return
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

	c.lastOrganizationsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastOrganizationsScrapeTimestampMetric.Collect(ch)

	c.lastOrganizationsScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastOrganizationsScrapeDurationSecondsMetric.Collect(ch)

	c.reportErrorMetric(false, ch)
}

func (c OrganizationsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.organizationInfoMetric.Describe(ch)
	c.organizationsTotalMetric.Describe(ch)
	c.lastOrganizationsScrapeErrorMetric.Describe(ch)
	c.lastOrganizationsScrapeTimestampMetric.Describe(ch)
	c.lastOrganizationsScrapeDurationSecondsMetric.Describe(ch)
}

func (c OrganizationsCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if errorHappend {
		errorMetric = float64(1)
	}

	c.lastOrganizationsScrapeErrorMetric.Set(errorMetric)
	c.lastOrganizationsScrapeErrorMetric.Collect(ch)
}
