package collectors

import (
	"time"

	"github.com/cloudfoundry/cf_exporter/v2/models"
	"github.com/prometheus/client_golang/prometheus"
)

type SecurityGroupsCollector struct {
	namespace                                     string
	environment                                   string
	deployment                                    string
	securityGroupInfoMetric                       *prometheus.GaugeVec
	securityGroupsScrapesTotalMetric              prometheus.Counter
	securityGroupsScrapeErrorsTotalMetric         prometheus.Counter
	lastSecurityGroupsScrapeErrorMetric           prometheus.Gauge
	lastSecurityGroupsScrapeTimestampMetric       prometheus.Gauge
	lastSecurityGroupsScrapeDurationSecondsMetric prometheus.Gauge
}

func NewSecurityGroupsCollector(
	namespace string,
	environment string,
	deployment string,
) *SecurityGroupsCollector {
	securityGroupInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "security_group",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Security Group information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"security_group_id", "security_group_name"},
	)

	securityGroupsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "security_groups_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Security Groups.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	securityGroupsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "security_groups_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Security Groups.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastSecurityGroupsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_security_groups_scrape_error",
			Help:        "Whether the last scrape of Security Groups metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastSecurityGroupsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_security_groups_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Security Groups metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastSecurityGroupsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_security_groups_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Security Groups metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &SecurityGroupsCollector{
		namespace:                                     namespace,
		environment:                                   environment,
		deployment:                                    deployment,
		securityGroupInfoMetric:                       securityGroupInfoMetric,
		securityGroupsScrapesTotalMetric:              securityGroupsScrapesTotalMetric,
		securityGroupsScrapeErrorsTotalMetric:         securityGroupsScrapeErrorsTotalMetric,
		lastSecurityGroupsScrapeErrorMetric:           lastSecurityGroupsScrapeErrorMetric,
		lastSecurityGroupsScrapeTimestampMetric:       lastSecurityGroupsScrapeTimestampMetric,
		lastSecurityGroupsScrapeDurationSecondsMetric: lastSecurityGroupsScrapeDurationSecondsMetric,
	}
}

func (c SecurityGroupsCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.Error != nil {
		errorMetric = float64(1)
		c.securityGroupsScrapeErrorsTotalMetric.Inc()
	} else {
		c.reportSecurityGroupsMetrics(objs, ch)
	}

	c.securityGroupsScrapeErrorsTotalMetric.Collect(ch)
	c.securityGroupsScrapesTotalMetric.Inc()
	c.securityGroupsScrapesTotalMetric.Collect(ch)
	c.lastSecurityGroupsScrapeErrorMetric.Set(errorMetric)
	c.lastSecurityGroupsScrapeErrorMetric.Collect(ch)
	c.lastSecurityGroupsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastSecurityGroupsScrapeTimestampMetric.Collect(ch)
	c.lastSecurityGroupsScrapeDurationSecondsMetric.Set(objs.Took)
	c.lastSecurityGroupsScrapeDurationSecondsMetric.Collect(ch)
}

func (c SecurityGroupsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.securityGroupInfoMetric.Describe(ch)
	c.securityGroupsScrapesTotalMetric.Describe(ch)
	c.securityGroupsScrapeErrorsTotalMetric.Describe(ch)
	c.lastSecurityGroupsScrapeErrorMetric.Describe(ch)
	c.lastSecurityGroupsScrapeTimestampMetric.Describe(ch)
	c.lastSecurityGroupsScrapeDurationSecondsMetric.Describe(ch)
}

func (c SecurityGroupsCollector) reportSecurityGroupsMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	c.securityGroupInfoMetric.Reset()

	for _, cSGroup := range objs.SecurityGroups {
		c.securityGroupInfoMetric.WithLabelValues(
			cSGroup.GUID,
			cSGroup.Name,
		).Set(float64(1))
	}
	c.securityGroupInfoMetric.Collect(ch)
}
