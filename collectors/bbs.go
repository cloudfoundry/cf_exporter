package collectors

import (
	"time"

	"github.com/cloudfoundry/cf_exporter/v2/models"
	"github.com/prometheus/client_golang/prometheus"
)

type BBSCollector struct {
	namespace                                    string
	environment                                  string
	deployment                                   string
	bbsActualLRPsScrapesTotalMetric              prometheus.Counter
	bbsActualLRPsScrapeErrorsTotalMetric         prometheus.Counter
	lastBBSActualLRPsScrapeErrorMetric           prometheus.Gauge
	lastBBSActualLRPsScrapeTimestampMetric       prometheus.Gauge
	lastBBSActualLRPsScrapeDurationSecondsMetric prometheus.Gauge
}

func NewBBSCollector(
	namespace string,
	environment string,
	deployment string,
) *BBSCollector {
	bbsActualLRPsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "bbs_actual_lrps_scrapes",
			Name:        "total",
			Help:        "Total number of BBS ActualLRPs scrapes.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	bbsActualLRPsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "bbs_actual_lrps_scrape_errors",
			Name:        "total",
			Help:        "Total number of BBS ActualLRPs scrape errors.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastBBSActualLRPsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_bbs_actual_lrps_scrape_error",
			Help:        "Whether the last BBS ActualLRPs scrape resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastBBSActualLRPsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_bbs_actual_lrps_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last BBS ActualLRPs scrape attempt.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastBBSActualLRPsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_bbs_actual_lrps_scrape_duration_seconds",
			Help:        "Duration of the last BBS ActualLRPs scrape attempt.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &BBSCollector{
		namespace:                                    namespace,
		environment:                                  environment,
		deployment:                                   deployment,
		bbsActualLRPsScrapesTotalMetric:              bbsActualLRPsScrapesTotalMetric,
		bbsActualLRPsScrapeErrorsTotalMetric:         bbsActualLRPsScrapeErrorsTotalMetric,
		lastBBSActualLRPsScrapeErrorMetric:           lastBBSActualLRPsScrapeErrorMetric,
		lastBBSActualLRPsScrapeTimestampMetric:       lastBBSActualLRPsScrapeTimestampMetric,
		lastBBSActualLRPsScrapeDurationSecondsMetric: lastBBSActualLRPsScrapeDurationSecondsMetric,
	}
}

func (c BBSCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.BBSActualLRPsError != nil {
		errorMetric = float64(1)
		c.bbsActualLRPsScrapeErrorsTotalMetric.Inc()
	}

	c.bbsActualLRPsScrapeErrorsTotalMetric.Collect(ch)
	c.bbsActualLRPsScrapesTotalMetric.Inc()
	c.bbsActualLRPsScrapesTotalMetric.Collect(ch)

	c.lastBBSActualLRPsScrapeErrorMetric.Set(errorMetric)
	c.lastBBSActualLRPsScrapeErrorMetric.Collect(ch)

	c.lastBBSActualLRPsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastBBSActualLRPsScrapeTimestampMetric.Collect(ch)

	c.lastBBSActualLRPsScrapeDurationSecondsMetric.Set(objs.Took)
	c.lastBBSActualLRPsScrapeDurationSecondsMetric.Collect(ch)
}

func (c BBSCollector) Describe(ch chan<- *prometheus.Desc) {
	c.bbsActualLRPsScrapesTotalMetric.Describe(ch)
	c.bbsActualLRPsScrapeErrorsTotalMetric.Describe(ch)
	c.lastBBSActualLRPsScrapeErrorMetric.Describe(ch)
	c.lastBBSActualLRPsScrapeTimestampMetric.Describe(ch)
	c.lastBBSActualLRPsScrapeDurationSecondsMetric.Describe(ch)
}
