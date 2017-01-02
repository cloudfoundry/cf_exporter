package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type SecurityGroupsCollector struct {
	namespace                               string
	cfClient                                *cfclient.Client
	securityGroupInfo                       *prometheus.GaugeVec
	securityGroupsTotal                     prometheus.Gauge
	lastSecurityGroupsScrapeError           prometheus.Gauge
	lastSecurityGroupsScrapeTimestamp       prometheus.Gauge
	lastSecurityGroupsScrapeDurationSeconds prometheus.Gauge
}

func NewSecurityGroupsCollector(namespace string, cfClient *cfclient.Client) *SecurityGroupsCollector {
	securityGroupInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "security_group",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Security Group information with a constant '1' value.",
		},
		[]string{"security_group_id", "security_group_name"},
	)

	securityGroupsTotal := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "security_groups",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Security Groups.",
		},
	)

	lastSecurityGroupsScrapeError := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_security_groups_scrape_error",
			Help:      "Whether the last scrape of Security Groups metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastSecurityGroupsScrapeTimestamp := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_security_groups_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Security Groups metrics from Cloud Foundry.",
		},
	)

	lastSecurityGroupsScrapeDurationSeconds := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_security_groups_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Security Groups metrics from Cloud Foundry.",
		},
	)

	return &SecurityGroupsCollector{
		namespace:                               namespace,
		cfClient:                                cfClient,
		securityGroupInfo:                       securityGroupInfo,
		securityGroupsTotal:                     securityGroupsTotal,
		lastSecurityGroupsScrapeError:           lastSecurityGroupsScrapeError,
		lastSecurityGroupsScrapeTimestamp:       lastSecurityGroupsScrapeTimestamp,
		lastSecurityGroupsScrapeDurationSeconds: lastSecurityGroupsScrapeDurationSeconds,
	}
}

func (c SecurityGroupsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	securityGroups, err := c.cfClient.ListSecGroups()
	if err != nil {
		log.Errorf("Error while listing security groups: %v", err)
		c.reportErrorMetric(true, ch)
		return
	}

	for _, securityGroup := range securityGroups {
		c.securityGroupInfo.WithLabelValues(
			securityGroup.Guid,
			securityGroup.Name,
		).Set(float64(1))
	}
	c.securityGroupInfo.Collect(ch)

	c.securityGroupsTotal.Set(float64(len(securityGroups)))
	c.securityGroupsTotal.Collect(ch)

	c.lastSecurityGroupsScrapeTimestamp.Set(float64(time.Now().Unix()))
	c.lastSecurityGroupsScrapeTimestamp.Collect(ch)

	c.lastSecurityGroupsScrapeDurationSeconds.Set(time.Since(begun).Seconds())
	c.lastSecurityGroupsScrapeDurationSeconds.Collect(ch)

	c.reportErrorMetric(false, ch)
}

func (c SecurityGroupsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.securityGroupInfo.Describe(ch)
	c.securityGroupsTotal.Describe(ch)
	c.lastSecurityGroupsScrapeError.Describe(ch)
	c.lastSecurityGroupsScrapeTimestamp.Describe(ch)
	c.lastSecurityGroupsScrapeDurationSeconds.Describe(ch)
}

func (c SecurityGroupsCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if errorHappend {
		errorMetric = float64(1)
	}

	c.lastSecurityGroupsScrapeError.Set(errorMetric)
	c.lastSecurityGroupsScrapeError.Collect(ch)

}
