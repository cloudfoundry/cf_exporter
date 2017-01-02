package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type OrganizationsCollector struct {
	namespace                              string
	cfClient                               *cfclient.Client
	organizationInfo                       *prometheus.GaugeVec
	organizationsTotal                     prometheus.Gauge
	lastOrganizationsScrapeError           prometheus.Gauge
	lastOrganizationsScrapeTimestamp       prometheus.Gauge
	lastOrganizationsScrapeDurationSeconds prometheus.Gauge
}

func NewOrganizationsCollector(namespace string, cfClient *cfclient.Client) *OrganizationsCollector {
	organizationInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "organization",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Organization information with a constant '1' value.",
		},
		[]string{"organization_id", "organization_name"},
	)

	organizationsTotal := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "organizations",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Organizations.",
		},
	)

	lastOrganizationsScrapeError := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_organizations_scrape_error",
			Help:      "Whether the last scrape of Organizations metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastOrganizationsScrapeTimestamp := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_organizations_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Organizations metrics from Cloud Foundry.",
		},
	)

	lastOrganizationsScrapeDurationSeconds := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_organizations_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Organizations metrics from Cloud Foundry.",
		},
	)

	return &OrganizationsCollector{
		namespace:                              namespace,
		cfClient:                               cfClient,
		organizationInfo:                       organizationInfo,
		organizationsTotal:                     organizationsTotal,
		lastOrganizationsScrapeError:           lastOrganizationsScrapeError,
		lastOrganizationsScrapeTimestamp:       lastOrganizationsScrapeTimestamp,
		lastOrganizationsScrapeDurationSeconds: lastOrganizationsScrapeDurationSeconds,
	}
}

func (c OrganizationsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	organizations, err := c.cfClient.ListOrgs()
	if err != nil {
		log.Errorf("Error while listing organizations: %v", err)
		c.reportErrorMetric(true, ch)
		return
	}

	for _, organization := range organizations {
		c.organizationInfo.WithLabelValues(
			organization.Guid,
			organization.Name,
		).Set(float64(1))
	}
	c.organizationInfo.Collect(ch)

	c.organizationsTotal.Set(float64(len(organizations)))
	c.organizationsTotal.Collect(ch)

	c.lastOrganizationsScrapeTimestamp.Set(float64(time.Now().Unix()))
	c.lastOrganizationsScrapeTimestamp.Collect(ch)

	c.lastOrganizationsScrapeDurationSeconds.Set(time.Since(begun).Seconds())
	c.lastOrganizationsScrapeDurationSeconds.Collect(ch)

	c.reportErrorMetric(false, ch)
}

func (c OrganizationsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.organizationInfo.Describe(ch)
	c.organizationsTotal.Describe(ch)
	c.lastOrganizationsScrapeError.Describe(ch)
	c.lastOrganizationsScrapeTimestamp.Describe(ch)
	c.lastOrganizationsScrapeDurationSeconds.Describe(ch)
}

func (c OrganizationsCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if errorHappend {
		errorMetric = float64(1)
	}

	c.lastOrganizationsScrapeError.Set(errorMetric)
	c.lastOrganizationsScrapeError.Collect(ch)
}
