package collectors

import (
	"time"

	"github.com/cloudfoundry/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
)

type IsolationSegmentsCollector struct {
	namespace                                        string
	environment                                      string
	deployment                                       string
	isolationSegmentInfoMetric                       *prometheus.GaugeVec
	isolationSegmentsScrapesTotalMetric              prometheus.Counter
	isolationSegmentsScrapeErrorsTotalMetric         prometheus.Counter
	lastIsolationSegmentsScrapeErrorMetric           prometheus.Gauge
	lastIsolationSegmentsScrapeTimestampMetric       prometheus.Gauge
	lastIsolationSegmentsScrapeDurationSecondsMetric prometheus.Gauge
}

func NewIsolationSegmentsCollector(
	namespace string,
	environment string,
	deployment string,
) *IsolationSegmentsCollector {
	isolationSegmentInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "isolation_segment",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Isolation Segment information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"isolation_segment_id", "isolation_segment_name"},
	)

	isolationSegmentsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "isolation_segments_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Isolation Segments.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	isolationSegmentsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "isolation_segments_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Isolation Segments.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastIsolationSegmentsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_isolation_segments_scrape_error",
			Help:        "Whether the last scrape of Isolation Segments metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastIsolationSegmentsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_isolation_segments_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Isolation Segments metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastIsolationSegmentsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_isolation_segments_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Isolation Segments metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &IsolationSegmentsCollector{
		namespace:                                        namespace,
		environment:                                      environment,
		deployment:                                       deployment,
		isolationSegmentInfoMetric:                       isolationSegmentInfoMetric,
		isolationSegmentsScrapesTotalMetric:              isolationSegmentsScrapesTotalMetric,
		isolationSegmentsScrapeErrorsTotalMetric:         isolationSegmentsScrapeErrorsTotalMetric,
		lastIsolationSegmentsScrapeErrorMetric:           lastIsolationSegmentsScrapeErrorMetric,
		lastIsolationSegmentsScrapeTimestampMetric:       lastIsolationSegmentsScrapeTimestampMetric,
		lastIsolationSegmentsScrapeDurationSecondsMetric: lastIsolationSegmentsScrapeDurationSecondsMetric,
	}
}

func (c IsolationSegmentsCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.Error != nil {
		errorMetric = float64(1)
		c.isolationSegmentsScrapeErrorsTotalMetric.Inc()
	} else {
		c.reportIsolationSegmentsMetrics(objs, ch)
	}

	c.isolationSegmentsScrapeErrorsTotalMetric.Collect(ch)
	c.isolationSegmentsScrapesTotalMetric.Inc()
	c.isolationSegmentsScrapesTotalMetric.Collect(ch)
	c.lastIsolationSegmentsScrapeErrorMetric.Set(errorMetric)
	c.lastIsolationSegmentsScrapeErrorMetric.Collect(ch)
	c.lastIsolationSegmentsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastIsolationSegmentsScrapeTimestampMetric.Collect(ch)
	c.lastIsolationSegmentsScrapeDurationSecondsMetric.Set(objs.Took)
	c.lastIsolationSegmentsScrapeDurationSecondsMetric.Collect(ch)
}

func (c IsolationSegmentsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.isolationSegmentInfoMetric.Describe(ch)
	c.isolationSegmentsScrapesTotalMetric.Describe(ch)
	c.isolationSegmentsScrapeErrorsTotalMetric.Describe(ch)
	c.lastIsolationSegmentsScrapeErrorMetric.Describe(ch)
	c.lastIsolationSegmentsScrapeTimestampMetric.Describe(ch)
	c.lastIsolationSegmentsScrapeDurationSecondsMetric.Describe(ch)
}

func (c IsolationSegmentsCollector) reportIsolationSegmentsMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	c.isolationSegmentInfoMetric.Reset()

	for _, s := range objs.Segments {
		c.isolationSegmentInfoMetric.WithLabelValues(
			s.GUID,
			s.Name,
		).Set(float64(1))
	}
	c.isolationSegmentInfoMetric.Collect(ch)
}
