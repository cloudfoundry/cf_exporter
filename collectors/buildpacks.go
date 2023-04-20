package collectors

import (
	"time"

	"github.com/bosh-prometheus/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
)

type BuildpacksCollector struct {
	namespace                                 string
	environment                               string
	deployment                                string
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
		buildpackInfoMetric:                       buildpackInfoMetric,
		buildpacksScrapesTotalMetric:              buildpacksScrapesTotalMetric,
		buildpacksScrapeErrorsTotalMetric:         buildpacksScrapeErrorsTotalMetric,
		lastBuildpacksScrapeErrorMetric:           lastBuildpacksScrapeErrorMetric,
		lastBuildpacksScrapeTimestampMetric:       lastBuildpacksScrapeTimestampMetric,
		lastBuildpacksScrapeDurationSecondsMetric: lastBuildpacksScrapeDurationSecondsMetric,
	}
}

func (c BuildpacksCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.Error != nil {
		errorMetric = float64(1)
		c.buildpacksScrapeErrorsTotalMetric.Inc()
	} else {
		c.reportBuildpacksMetrics(objs, ch)
	}

	c.buildpacksScrapeErrorsTotalMetric.Collect(ch)
	c.buildpacksScrapesTotalMetric.Inc()
	c.buildpacksScrapesTotalMetric.Collect(ch)
	c.lastBuildpacksScrapeErrorMetric.Set(errorMetric)
	c.lastBuildpacksScrapeErrorMetric.Collect(ch)
	c.lastBuildpacksScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastBuildpacksScrapeTimestampMetric.Collect(ch)
	c.lastBuildpacksScrapeDurationSecondsMetric.Set(objs.Took)
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

func (c BuildpacksCollector) reportBuildpacksMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	c.buildpackInfoMetric.Reset()

	for _, buildpack := range objs.Buildpacks {
		c.buildpackInfoMetric.WithLabelValues(
			buildpack.GUID,
			buildpack.Name,
			buildpack.Stack,
			buildpack.Filename,
		).Set(float64(1))
	}

	c.buildpackInfoMetric.Collect(ch)
}
