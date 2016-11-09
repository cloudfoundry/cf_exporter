package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type OrganizationsCollector struct {
	namespace                                  string
	cfClient                                   *cfclient.Client
	organizationInfoDesc                       *prometheus.Desc
	organizationsTotalDesc                     *prometheus.Desc
	lastOrganizationsScrapeError               *prometheus.Desc
	lastOrganizationsScrapeTimestampDesc       *prometheus.Desc
	lastOrganizationsScrapeDurationSecondsDesc *prometheus.Desc
}

func NewOrganizationsCollector(namespace string, cfClient *cfclient.Client) *OrganizationsCollector {
	organizationInfoDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "organization", "info"),
		"Labeled Cloud Foundry Organization information with a constant '1' value.",
		[]string{"organization_id", "organization_name"},
		nil,
	)

	organizationsTotalDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "organizations", "total"),
		"Total number of Cloud Foundry Organizations.",
		[]string{},
		nil,
	)

	lastOrganizationsScrapeError := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_organizations_scrape_error"),
		"Whether the last scrape of Organizations metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		[]string{},
		nil,
	)

	lastOrganizationsScrapeTimestampDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_organizations_scrape_timestamp"),
		"Number of seconds since 1970 since last scrape of Organizations metrics from Cloud Foundry.",
		[]string{},
		nil,
	)

	lastOrganizationsScrapeDurationSecondsDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_organizations_scrape_duration_seconds"),
		"Duration of the last scrape of Organizations metrics from Cloud Foundry.",
		[]string{},
		nil,
	)

	return &OrganizationsCollector{
		namespace:                                  namespace,
		cfClient:                                   cfClient,
		organizationInfoDesc:                       organizationInfoDesc,
		organizationsTotalDesc:                     organizationsTotalDesc,
		lastOrganizationsScrapeError:               lastOrganizationsScrapeError,
		lastOrganizationsScrapeTimestampDesc:       lastOrganizationsScrapeTimestampDesc,
		lastOrganizationsScrapeDurationSecondsDesc: lastOrganizationsScrapeDurationSecondsDesc,
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
		ch <- prometheus.MustNewConstMetric(
			c.organizationInfoDesc,
			prometheus.GaugeValue,
			float64(1),
			organization.Guid,
			organization.Name,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.organizationsTotalDesc,
		prometheus.GaugeValue,
		float64(len(organizations)),
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastOrganizationsScrapeTimestampDesc,
		prometheus.GaugeValue,
		float64(time.Now().Unix()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastOrganizationsScrapeDurationSecondsDesc,
		prometheus.GaugeValue,
		time.Since(begun).Seconds(),
	)

	c.reportErrorMetric(false, ch)
}

func (c OrganizationsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.organizationInfoDesc
	ch <- c.organizationsTotalDesc
	ch <- c.lastOrganizationsScrapeError
	ch <- c.lastOrganizationsScrapeTimestampDesc
	ch <- c.lastOrganizationsScrapeDurationSecondsDesc
}

func (c OrganizationsCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	error_metric := float64(0)
	if errorHappend {
		error_metric = float64(1)
	}

	ch <- prometheus.MustNewConstMetric(
		c.lastOrganizationsScrapeError,
		prometheus.GaugeValue,
		error_metric,
	)
}
