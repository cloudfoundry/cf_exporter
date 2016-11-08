package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ApplicationsCollector struct {
	namespace                                 string
	cfClient                                  *cfclient.Client
	applicationInfoDesc                       *prometheus.Desc
	applicationsTotalDesc                     *prometheus.Desc
	lastApplicationsScrapeTimestampDesc       *prometheus.Desc
	lastApplicationsScrapeDurationSecondsDesc *prometheus.Desc
}

func NewApplicationsCollector(namespace string, cfClient *cfclient.Client) *ApplicationsCollector {
	applicationInfoDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "application", "info"),
		"Cloud Foundry Application information.",
		[]string{"application_id", "application_name", "space_id", "space_name", "organization_id", "organization_name"},
		nil,
	)

	applicationsTotalDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "applications", "total"),
		"Total number of Cloud Foundry Applications.",
		[]string{},
		nil,
	)

	lastApplicationsScrapeTimestampDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_applications_scrape_timestamp"),
		"Number of seconds since 1970 since last scrape of Applications metrics from Cloud Foundry.",
		[]string{},
		nil,
	)

	lastApplicationsScrapeDurationSecondsDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_applications_scrape_duration_seconds"),
		"Duration of the last scrape of Applications metrics from Cloud Foundry.",
		[]string{},
		nil,
	)

	return &ApplicationsCollector{
		namespace:                                 namespace,
		cfClient:                                  cfClient,
		applicationsTotalDesc:                     applicationsTotalDesc,
		applicationInfoDesc:                       applicationInfoDesc,
		lastApplicationsScrapeTimestampDesc:       lastApplicationsScrapeTimestampDesc,
		lastApplicationsScrapeDurationSecondsDesc: lastApplicationsScrapeDurationSecondsDesc,
	}
}

func (c ApplicationsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	applications, err := c.cfClient.ListApps()
	if err != nil {
		log.Errorf("Error while listing applications: %v", err)
		return
	}

	for _, application := range applications {
		ch <- prometheus.MustNewConstMetric(
			c.applicationInfoDesc,
			prometheus.GaugeValue,
			float64(1),
			application.Guid,
			application.Name,
			application.SpaceData.Entity.Guid,
			application.SpaceData.Entity.Name,
			application.SpaceData.Entity.OrgData.Entity.Guid,
			application.SpaceData.Entity.OrgData.Entity.Name,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.applicationsTotalDesc,
		prometheus.GaugeValue,
		float64(len(applications)),
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastApplicationsScrapeTimestampDesc,
		prometheus.GaugeValue,
		float64(time.Now().Unix()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastApplicationsScrapeDurationSecondsDesc,
		prometheus.GaugeValue,
		time.Since(begun).Seconds(),
	)
}

func (c ApplicationsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.applicationInfoDesc
	ch <- c.applicationsTotalDesc
	ch <- c.lastApplicationsScrapeTimestampDesc
	ch <- c.lastApplicationsScrapeDurationSecondsDesc
}
