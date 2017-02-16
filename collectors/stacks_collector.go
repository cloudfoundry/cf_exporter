package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type StacksCollector struct {
	namespace                             string
	environment                           string
	deployment                            string
	cfClient                              *cfclient.Client
	stackInfoMetric                       *prometheus.GaugeVec
	stacksScrapesTotalMetric              prometheus.Counter
	stacksScrapeErrorsTotalMetric         prometheus.Counter
	lastStacksScrapeErrorMetric           prometheus.Gauge
	lastStacksScrapeTimestampMetric       prometheus.Gauge
	lastStacksScrapeDurationSecondsMetric prometheus.Gauge
}

func NewStacksCollector(
	namespace string,
	environment string,
	deployment string,
	cfClient *cfclient.Client,
) *StacksCollector {
	stackInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "stack",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Stack information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"stack_id", "stack_name"},
	)

	stacksScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "stacks_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Stacks.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	stacksScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "stacks_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Stacks.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastStacksScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_stacks_scrape_error",
			Help:        "Whether the last scrape of Stacks metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastStacksScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_stacks_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Stacks metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastStacksScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_stacks_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Stacks metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &StacksCollector{
		namespace:                             namespace,
		environment:                           environment,
		deployment:                            deployment,
		cfClient:                              cfClient,
		stackInfoMetric:                       stackInfoMetric,
		stacksScrapesTotalMetric:              stacksScrapesTotalMetric,
		stacksScrapeErrorsTotalMetric:         stacksScrapeErrorsTotalMetric,
		lastStacksScrapeErrorMetric:           lastStacksScrapeErrorMetric,
		lastStacksScrapeTimestampMetric:       lastStacksScrapeTimestampMetric,
		lastStacksScrapeDurationSecondsMetric: lastStacksScrapeDurationSecondsMetric,
	}
}

func (c StacksCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportStacksMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.stacksScrapeErrorsTotalMetric.Inc()
	}
	c.stacksScrapeErrorsTotalMetric.Collect(ch)

	c.stacksScrapesTotalMetric.Inc()
	c.stacksScrapesTotalMetric.Collect(ch)

	c.lastStacksScrapeErrorMetric.Set(errorMetric)
	c.lastStacksScrapeErrorMetric.Collect(ch)

	c.lastStacksScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastStacksScrapeTimestampMetric.Collect(ch)

	c.lastStacksScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastStacksScrapeDurationSecondsMetric.Collect(ch)
}

func (c StacksCollector) Describe(ch chan<- *prometheus.Desc) {
	c.stackInfoMetric.Describe(ch)
	c.stacksScrapesTotalMetric.Describe(ch)
	c.stacksScrapeErrorsTotalMetric.Describe(ch)
	c.lastStacksScrapeErrorMetric.Describe(ch)
	c.lastStacksScrapeTimestampMetric.Describe(ch)
	c.lastStacksScrapeDurationSecondsMetric.Describe(ch)
}

func (c StacksCollector) reportStacksMetrics(ch chan<- prometheus.Metric) error {
	c.stackInfoMetric.Reset()

	stacks, err := c.cfClient.ListStacks()
	if err != nil {
		log.Errorf("Error while listing stacks: %v", err)
		return err
	}

	for _, stack := range stacks {
		c.stackInfoMetric.WithLabelValues(
			stack.Guid,
			stack.Name,
		).Set(float64(1))
	}

	c.stackInfoMetric.Collect(ch)

	return nil
}
