package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type SecurityGroupsCollector struct {
	namespace                                     string
	cfClient                                      *cfclient.Client
	securityGroupInfoMetric                       *prometheus.GaugeVec
	securityGroupsTotalMetric                     prometheus.Gauge
	securityGroupsScrapesTotalMetric              prometheus.Counter
	securityGroupsScrapeErrorsTotalMetric         prometheus.Counter
	lastSecurityGroupsScrapeErrorMetric           prometheus.Gauge
	lastSecurityGroupsScrapeTimestampMetric       prometheus.Gauge
	lastSecurityGroupsScrapeDurationSecondsMetric prometheus.Gauge
}

func NewSecurityGroupsCollector(namespace string, cfClient *cfclient.Client) *SecurityGroupsCollector {
	securityGroupInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "security_group",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Security Group information with a constant '1' value.",
		},
		[]string{"security_group_id", "security_group_name"},
	)

	securityGroupsTotalMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "security_groups",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Security Groups.",
		},
	)

	securityGroupsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "security_groups_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Cloud Foundry Security Groups.",
		},
	)

	securityGroupsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "security_groups_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrape error of Cloud Foundry Security Groups.",
		},
	)

	lastSecurityGroupsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_security_groups_scrape_error",
			Help:      "Whether the last scrape of Security Groups metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastSecurityGroupsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_security_groups_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Security Groups metrics from Cloud Foundry.",
		},
	)

	lastSecurityGroupsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_security_groups_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Security Groups metrics from Cloud Foundry.",
		},
	)

	return &SecurityGroupsCollector{
		namespace:                                     namespace,
		cfClient:                                      cfClient,
		securityGroupInfoMetric:                       securityGroupInfoMetric,
		securityGroupsTotalMetric:                     securityGroupsTotalMetric,
		securityGroupsScrapesTotalMetric:              securityGroupsScrapesTotalMetric,
		securityGroupsScrapeErrorsTotalMetric:         securityGroupsScrapeErrorsTotalMetric,
		lastSecurityGroupsScrapeErrorMetric:           lastSecurityGroupsScrapeErrorMetric,
		lastSecurityGroupsScrapeTimestampMetric:       lastSecurityGroupsScrapeTimestampMetric,
		lastSecurityGroupsScrapeDurationSecondsMetric: lastSecurityGroupsScrapeDurationSecondsMetric,
	}
}

func (c SecurityGroupsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportSecurityGroupsMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.securityGroupsScrapeErrorsTotalMetric.Inc()
	}

	c.securityGroupsScrapesTotalMetric.Inc()
	c.securityGroupsScrapesTotalMetric.Collect(ch)

	c.securityGroupsScrapeErrorsTotalMetric.Collect(ch)

	c.lastSecurityGroupsScrapeErrorMetric.Set(errorMetric)
	c.lastSecurityGroupsScrapeErrorMetric.Collect(ch)

	c.lastSecurityGroupsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastSecurityGroupsScrapeTimestampMetric.Collect(ch)

	c.lastSecurityGroupsScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastSecurityGroupsScrapeDurationSecondsMetric.Collect(ch)
}

func (c SecurityGroupsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.securityGroupInfoMetric.Describe(ch)
	c.securityGroupsTotalMetric.Describe(ch)
	c.securityGroupsScrapesTotalMetric.Describe(ch)
	c.securityGroupsScrapeErrorsTotalMetric.Describe(ch)
	c.lastSecurityGroupsScrapeErrorMetric.Describe(ch)
	c.lastSecurityGroupsScrapeTimestampMetric.Describe(ch)
	c.lastSecurityGroupsScrapeDurationSecondsMetric.Describe(ch)
}

func (c SecurityGroupsCollector) reportSecurityGroupsMetrics(ch chan<- prometheus.Metric) error {
	c.securityGroupInfoMetric.Reset()

	securityGroups, err := c.cfClient.ListSecGroups()
	if err != nil {
		log.Errorf("Error while listing security groups: %v", err)
		return err
	}

	for _, securityGroup := range securityGroups {
		c.securityGroupInfoMetric.WithLabelValues(
			securityGroup.Guid,
			securityGroup.Name,
		).Set(float64(1))
	}

	c.securityGroupInfoMetric.Collect(ch)

	c.securityGroupsTotalMetric.Set(float64(len(securityGroups)))
	c.securityGroupsTotalMetric.Collect(ch)

	return nil
}
