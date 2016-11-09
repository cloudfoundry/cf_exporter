package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type SpacesCollector struct {
	namespace                           string
	cfClient                            *cfclient.Client
	spaceInfoDesc                       *prometheus.Desc
	spacesTotalDesc                     *prometheus.Desc
	lastSpacesScrapeError               *prometheus.Desc
	lastSpacesScrapeTimestampDesc       *prometheus.Desc
	lastSpacesScrapeDurationSecondsDesc *prometheus.Desc
}

func NewSpacesCollector(namespace string, cfClient *cfclient.Client) *SpacesCollector {
	spaceInfoDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "space", "info"),
		"Labeled Cloud Foundry Space information with a constant '1' value.",
		[]string{"space_id", "space_name"},
		nil,
	)

	spacesTotalDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "spaces", "total"),
		"Total number of Cloud Foundry Spaces.",
		[]string{},
		nil,
	)

	lastSpacesScrapeError := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_spaces_scrape_error"),
		"Whether the last scrape of Spaces metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		[]string{},
		nil,
	)

	lastSpacesScrapeTimestampDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_spaces_scrape_timestamp"),
		"Number of seconds since 1970 since last scrape of Spaces metrics from Cloud Foundry.",
		[]string{},
		nil,
	)

	lastSpacesScrapeDurationSecondsDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_spaces_scrape_duration_seconds"),
		"Duration of the last scrape of Spaces metrics from Cloud Foundry.",
		[]string{},
		nil,
	)

	return &SpacesCollector{
		namespace:                           namespace,
		cfClient:                            cfClient,
		spaceInfoDesc:                       spaceInfoDesc,
		spacesTotalDesc:                     spacesTotalDesc,
		lastSpacesScrapeError:               lastSpacesScrapeError,
		lastSpacesScrapeTimestampDesc:       lastSpacesScrapeTimestampDesc,
		lastSpacesScrapeDurationSecondsDesc: lastSpacesScrapeDurationSecondsDesc,
	}
}

func (c SpacesCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	spaces, err := c.cfClient.ListSpaces()
	if err != nil {
		log.Errorf("Error while listing spaces: %v", err)
		c.reportErrorMetric(true, ch)
		return
	}

	for _, space := range spaces {
		ch <- prometheus.MustNewConstMetric(
			c.spaceInfoDesc,
			prometheus.GaugeValue,
			float64(1),
			space.Guid,
			space.Name,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.spacesTotalDesc,
		prometheus.GaugeValue,
		float64(len(spaces)),
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastSpacesScrapeTimestampDesc,
		prometheus.GaugeValue,
		float64(time.Now().Unix()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastSpacesScrapeDurationSecondsDesc,
		prometheus.GaugeValue,
		time.Since(begun).Seconds(),
	)

	c.reportErrorMetric(false, ch)
}

func (c SpacesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.spaceInfoDesc
	ch <- c.spacesTotalDesc
	ch <- c.lastSpacesScrapeError
	ch <- c.lastSpacesScrapeTimestampDesc
	ch <- c.lastSpacesScrapeDurationSecondsDesc
}

func (c SpacesCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	error_metric := float64(0)
	if errorHappend {
		error_metric = float64(1)
	}

	ch <- prometheus.MustNewConstMetric(
		c.lastSpacesScrapeError,
		prometheus.GaugeValue,
		error_metric,
	)
}
