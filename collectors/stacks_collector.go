package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type StacksCollector struct {
	namespace                             string
	deploymentName                        string
	cfClient                              *cfclient.Client
	stackInfoMetric                       *prometheus.GaugeVec
	stacksScrapesTotalMetric              *prometheus.CounterVec
	stacksScrapeErrorsTotalMetric         *prometheus.CounterVec
	lastStacksScrapeErrorMetric           *prometheus.GaugeVec
	lastStacksScrapeTimestampMetric       *prometheus.GaugeVec
	lastStacksScrapeDurationSecondsMetric *prometheus.GaugeVec
}

func NewStacksCollector(namespace string, deploymentName string, cfClient *cfclient.Client) *StacksCollector {
	stackInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "stack",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Stack information with a constant '1' value.",
		},
		[]string{"deployment", "stack_id", "stack_name"},
	)

	stacksScrapesTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "stacks_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Cloud Foundry Stacks.",
		},
		[]string{"deployment"},
	)

	stacksScrapeErrorsTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "stacks_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrape error of Cloud Foundry Stacks.",
		},
		[]string{"deployment"},
	)

	lastStacksScrapeErrorMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_stacks_scrape_error",
			Help:      "Whether the last scrape of Stacks metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
		[]string{"deployment"},
	)

	lastStacksScrapeTimestampMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_stacks_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Stacks metrics from Cloud Foundry.",
		},
		[]string{"deployment"},
	)

	lastStacksScrapeDurationSecondsMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_stacks_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Stacks metrics from Cloud Foundry.",
		},
		[]string{"deployment"},
	)

	return &StacksCollector{
		namespace:                             namespace,
		deploymentName:                        deploymentName,
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
		c.stacksScrapeErrorsTotalMetric.WithLabelValues(c.deploymentName).Inc()
	}
	c.stacksScrapeErrorsTotalMetric.Collect(ch)

	c.stacksScrapesTotalMetric.WithLabelValues(c.deploymentName).Inc()
	c.stacksScrapesTotalMetric.Collect(ch)

	c.lastStacksScrapeErrorMetric.WithLabelValues(c.deploymentName).Set(errorMetric)
	c.lastStacksScrapeErrorMetric.Collect(ch)

	c.lastStacksScrapeTimestampMetric.WithLabelValues(c.deploymentName).Set(float64(time.Now().Unix()))
	c.lastStacksScrapeTimestampMetric.Collect(ch)

	c.lastStacksScrapeDurationSecondsMetric.WithLabelValues(c.deploymentName).Set(time.Since(begun).Seconds())
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
			c.deploymentName,
			stack.Guid,
			stack.Name,
		).Set(float64(1))
	}

	c.stackInfoMetric.Collect(ch)

	return nil
}
