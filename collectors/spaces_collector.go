package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type SpacesCollector struct {
	namespace                       string
	cfClient                        *cfclient.Client
	spaceInfo                       *prometheus.GaugeVec
	spacesTotal                     prometheus.Gauge
	lastSpacesScrapeError           prometheus.Gauge
	lastSpacesScrapeTimestamp       prometheus.Gauge
	lastSpacesScrapeDurationSeconds prometheus.Gauge
}

func NewSpacesCollector(namespace string, cfClient *cfclient.Client) *SpacesCollector {
	spaceInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "space",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Space information with a constant '1' value.",
		},
		[]string{"space_id", "space_name"},
	)

	spacesTotal := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "spaces",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Spaces.",
		},
	)

	lastSpacesScrapeError := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_spaces_scrape_error",
			Help:      "Whether the last scrape of Spaces metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastSpacesScrapeTimestamp := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_spaces_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Spaces metrics from Cloud Foundry.",
		},
	)

	lastSpacesScrapeDurationSeconds := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_spaces_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Spaces metrics from Cloud Foundry.",
		},
	)

	return &SpacesCollector{
		namespace:                       namespace,
		cfClient:                        cfClient,
		spaceInfo:                       spaceInfo,
		spacesTotal:                     spacesTotal,
		lastSpacesScrapeError:           lastSpacesScrapeError,
		lastSpacesScrapeTimestamp:       lastSpacesScrapeTimestamp,
		lastSpacesScrapeDurationSeconds: lastSpacesScrapeDurationSeconds,
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
		c.spaceInfo.WithLabelValues(
			space.Guid,
			space.Name,
		).Set(float64(1))
	}
	c.spaceInfo.Collect(ch)

	c.spacesTotal.Set(float64(len(spaces)))
	c.spacesTotal.Collect(ch)

	c.lastSpacesScrapeTimestamp.Set(float64(time.Now().Unix()))
	c.lastSpacesScrapeTimestamp.Collect(ch)

	c.lastSpacesScrapeDurationSeconds.Set(time.Since(begun).Seconds())
	c.lastSpacesScrapeDurationSeconds.Collect(ch)

	c.reportErrorMetric(false, ch)
}

func (c SpacesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.spaceInfo.Describe(ch)
	c.spacesTotal.Describe(ch)
	c.lastSpacesScrapeError.Describe(ch)
	c.lastSpacesScrapeTimestamp.Describe(ch)
	c.lastSpacesScrapeDurationSeconds.Describe(ch)
}

func (c SpacesCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if errorHappend {
		errorMetric = float64(1)
	}

	c.lastSpacesScrapeError.Set(errorMetric)
	c.lastSpacesScrapeError.Collect(ch)
}
