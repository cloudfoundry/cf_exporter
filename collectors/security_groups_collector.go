package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type SecurityGroupsCollector struct {
	namespace                                     string
	environment                                   string
	deploymentName                                string
	cfClient                                      *cfclient.Client
	securityGroupInfoMetric                       *prometheus.GaugeVec
	securityGroupsScrapesTotalMetric              *prometheus.CounterVec
	securityGroupsScrapeErrorsTotalMetric         *prometheus.CounterVec
	lastSecurityGroupsScrapeErrorMetric           *prometheus.GaugeVec
	lastSecurityGroupsScrapeTimestampMetric       *prometheus.GaugeVec
	lastSecurityGroupsScrapeDurationSecondsMetric *prometheus.GaugeVec
}

func NewSecurityGroupsCollector(
	namespace string,
	environment string,
	deploymentName string,
	cfClient *cfclient.Client,
) *SecurityGroupsCollector {
	securityGroupInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "security_group",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Security Group information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment", "security_group_id", "security_group_name"},
	)

	securityGroupsScrapesTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "security_groups_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Security Groups.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	securityGroupsScrapeErrorsTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "security_groups_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Security Groups.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	lastSecurityGroupsScrapeErrorMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_security_groups_scrape_error",
			Help:        "Whether the last scrape of Security Groups metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	lastSecurityGroupsScrapeTimestampMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_security_groups_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Security Groups metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	lastSecurityGroupsScrapeDurationSecondsMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_security_groups_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Security Groups metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	return &SecurityGroupsCollector{
		namespace:                                     namespace,
		environment:                                   environment,
		deploymentName:                                deploymentName,
		cfClient:                                      cfClient,
		securityGroupInfoMetric:                       securityGroupInfoMetric,
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
		c.securityGroupsScrapeErrorsTotalMetric.WithLabelValues(c.deploymentName).Inc()
	}
	c.securityGroupsScrapeErrorsTotalMetric.Collect(ch)

	c.securityGroupsScrapesTotalMetric.WithLabelValues(c.deploymentName).Inc()
	c.securityGroupsScrapesTotalMetric.Collect(ch)

	c.lastSecurityGroupsScrapeErrorMetric.WithLabelValues(c.deploymentName).Set(errorMetric)
	c.lastSecurityGroupsScrapeErrorMetric.Collect(ch)

	c.lastSecurityGroupsScrapeTimestampMetric.WithLabelValues(c.deploymentName).Set(float64(time.Now().Unix()))
	c.lastSecurityGroupsScrapeTimestampMetric.Collect(ch)

	c.lastSecurityGroupsScrapeDurationSecondsMetric.WithLabelValues(c.deploymentName).Set(time.Since(begun).Seconds())
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

func (c SecurityGroupsCollector) reportSecurityGroupsMetrics(ch chan<- prometheus.Metric) error {
	c.securityGroupInfoMetric.Reset()

	securityGroups, err := c.cfClient.ListSecGroups()
	if err != nil {
		log.Errorf("Error while listing security groups: %v", err)
		return err
	}

	for _, securityGroup := range securityGroups {
		c.securityGroupInfoMetric.WithLabelValues(
			c.deploymentName,
			securityGroup.Guid,
			securityGroup.Name,
		).Set(float64(1))
	}

	c.securityGroupInfoMetric.Collect(ch)

	return nil
}
