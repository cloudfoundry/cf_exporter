package collectors

import (
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type BuildpacksCollector struct {
	namespace                                 string
	environment                               string
	deployment                                string
	cfClient                                  *cfclient.Client
	buildpackInfoMetric                       *prometheus.GaugeVec
	buildpacksScrapesTotalMetric              prometheus.Counter
	buildpacksScrapeErrorsTotalMetric         prometheus.Counter
	lastBuildpacksScrapeErrorMetric           prometheus.Gauge
	lastBuildpacksScrapeTimestampMetric       prometheus.Gauge
	lastBuildpacksScrapeDurationSecondsMetric prometheus.Gauge
}

func NewBuildpacksCollector(
	namespace string,
	environment string,
	deployment string,
	cfClient *cfclient.Client,
) *BuildpacksCollector {
	buildpackInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "buildpack",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Buildpack information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"buildpack_id", "buildpack_name", "buildpack_stack", "buildpack_filename"},
	)

	buildpacksScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "buildpacks_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Buildpacks.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	buildpacksScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "buildpacks_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Buildpacks.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastBuildpacksScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_buildpacks_scrape_error",
			Help:        "Whether the last scrape of Buildpacks metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastBuildpacksScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_buildpacks_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Buildpacks metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastBuildpacksScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_buildpacks_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Buildpacks metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &BuildpacksCollector{
		namespace:                                 namespace,
		environment:                               environment,
		deployment:                                deployment,
		cfClient:                                  cfClient,
		buildpackInfoMetric:                       buildpackInfoMetric,
		buildpacksScrapesTotalMetric:              buildpacksScrapesTotalMetric,
		buildpacksScrapeErrorsTotalMetric:         buildpacksScrapeErrorsTotalMetric,
		lastBuildpacksScrapeErrorMetric:           lastBuildpacksScrapeErrorMetric,
		lastBuildpacksScrapeTimestampMetric:       lastBuildpacksScrapeTimestampMetric,
		lastBuildpacksScrapeDurationSecondsMetric: lastBuildpacksScrapeDurationSecondsMetric,
	}
}

func (c BuildpacksCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportBuildpacksMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.buildpacksScrapeErrorsTotalMetric.Inc()
	}
	c.buildpacksScrapeErrorsTotalMetric.Collect(ch)

	c.buildpacksScrapesTotalMetric.Inc()
	c.buildpacksScrapesTotalMetric.Collect(ch)

	c.lastBuildpacksScrapeErrorMetric.Set(errorMetric)
	c.lastBuildpacksScrapeErrorMetric.Collect(ch)

	c.lastBuildpacksScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastBuildpacksScrapeTimestampMetric.Collect(ch)

	c.lastBuildpacksScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastBuildpacksScrapeDurationSecondsMetric.Collect(ch)
}

func (c BuildpacksCollector) Describe(ch chan<- *prometheus.Desc) {
	c.buildpackInfoMetric.Describe(ch)
	c.buildpacksScrapesTotalMetric.Describe(ch)
	c.buildpacksScrapeErrorsTotalMetric.Describe(ch)
	c.lastBuildpacksScrapeErrorMetric.Describe(ch)
	c.lastBuildpacksScrapeTimestampMetric.Describe(ch)
	c.lastBuildpacksScrapeDurationSecondsMetric.Describe(ch)
}

func (c BuildpacksCollector) reportBuildpacksMetrics(ch chan<- prometheus.Metric) error {
	c.buildpackInfoMetric.Reset()

	buildpacks, err := c.cfClient.ListBuildpacks()
	if err != nil {
		log.Errorf("Error while listing buildpacks: %v", err)
		return err
	}

	for _, buildpack := range buildpacks {
		c.buildpackInfoMetric.WithLabelValues(
			buildpack.Guid,
			buildpack.Name,
			buildpack.Stack,
			buildpack.Filename,
		).Set(float64(1))
	}

	c.buildpackInfoMetric.Collect(ch)

	return nil
}
