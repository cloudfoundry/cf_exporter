package collectors

import (
    "time"

    "github.com/bosh-prometheus/cf_exporter/models"
    "github.com/prometheus/client_golang/prometheus"
)

type DomainsCollector struct {
    namespace                                   string
    environment                                 string
    deployment                                  string
    domainInfoMetric                            *prometheus.GaugeVec
    domainInfoScrapesTotalMetric                prometheus.Counter
    domainInfoScrapeErrorsTotalMetric           prometheus.Counter
    lastDomainInfoScrapeErrorMetric             prometheus.Gauge
    lastDomainInfoScrapeTimestampMetric         prometheus.Gauge
    lastDomainInfoScrapeDurationSecondsMetric   prometheus.Gauge
}

func NewDomainsCollector(namespace string, environment string, deployment string) *DomainsCollector {
    domainInfoMetric := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace:   namespace,
            Subsystem:   "domain",
            Name:        "info",
            Help:        "Cloud Foundry domains, labeled by domain ID, name, whether it is internal, and supported protocol. Metric value is set to 1.",
            ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
        },
        []string{"domain_id", "domain_name", "internal", "protocol"},
    )

    domainInfoScrapesTotalMetric := prometheus.NewCounter(
        prometheus.CounterOpts{
            Namespace:   namespace,
            Subsystem:   "domain_info_scrapes",
            Name:        "total",
            Help:        "Total number of scrapes for Cloud Foundry Domains.",
            ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
        },
    )

    domainInfoScrapeErrorsTotalMetric := prometheus.NewCounter(
        prometheus.CounterOpts{
            Namespace:   namespace,
            Subsystem:   "domain_info_scrape_errors",
            Name:        "total",
            Help:        "Total number of scrape errors of Cloud Foundry Domains.",
            ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
        },
    )

    lastDomainInfoScrapeErrorMetric := prometheus.NewGauge(
        prometheus.GaugeOpts{
            Namespace:   namespace,
            Subsystem:   "",
            Name:        "last_domain_info_scrape_error",
            Help:        "Whether the last scrape of Domains metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
            ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
        },
    )

    lastDomainInfoScrapeTimestampMetric := prometheus.NewGauge(
        prometheus.GaugeOpts{
            Namespace:   namespace,
            Subsystem:   "",
            Name:        "last_domain_info_scrape_timestamp",
            Help:        "Number of seconds since 1970 since last scrape of Domains metrics from Cloud Foundry.",
            ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
        },
    )

    lastDomainInfoScrapeDurationSecondsMetric := prometheus.NewGauge(
        prometheus.GaugeOpts{
            Namespace:   namespace,
            Subsystem:   "",
            Name:        "last_domain_info_scrape_duration_seconds",
            Help:        "Duration of the last scrape of Domains metrics from Cloud Foundry.",
            ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
        },
    )

    return &DomainsCollector{
        namespace:                              namespace,
        environment:                            environment,
        deployment:                             deployment,
        domainInfoMetric:                       domainInfoMetric,
        domainInfoScrapesTotalMetric:              domainInfoScrapesTotalMetric,
        domainInfoScrapeErrorsTotalMetric:         domainInfoScrapeErrorsTotalMetric,
        lastDomainInfoScrapeErrorMetric:           lastDomainInfoScrapeErrorMetric,
        lastDomainInfoScrapeTimestampMetric:       lastDomainInfoScrapeTimestampMetric,
        lastDomainInfoScrapeDurationSecondsMetric: lastDomainInfoScrapeDurationSecondsMetric,
    }
}

func (c *DomainsCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
    errorMetric := float64(0)
    if objs.Error != nil {
        errorMetric = float64(1)
        c.domainInfoScrapeErrorsTotalMetric.Inc()
    } else {
        c.reportDomainsMetrics(objs, ch)
    }

    c.domainInfoScrapeErrorsTotalMetric.Collect(ch)
    c.domainInfoScrapesTotalMetric.Inc()
    c.domainInfoScrapesTotalMetric.Collect(ch)
    c.lastDomainInfoScrapeErrorMetric.Set(errorMetric)
    c.lastDomainInfoScrapeErrorMetric.Collect(ch)
    c.lastDomainInfoScrapeTimestampMetric.Set(float64(time.Now().Unix()))
    c.lastDomainInfoScrapeTimestampMetric.Collect(ch)
    c.lastDomainInfoScrapeDurationSecondsMetric.Set(objs.Took)
    c.lastDomainInfoScrapeDurationSecondsMetric.Collect(ch)
}

func (c *DomainsCollector) Describe(ch chan<- *prometheus.Desc) {
    c.domainInfoMetric.Describe(ch)
    c.domainInfoScrapesTotalMetric.Describe(ch)
    c.domainInfoScrapeErrorsTotalMetric.Describe(ch)
    c.lastDomainInfoScrapeErrorMetric.Describe(ch)
    c.lastDomainInfoScrapeTimestampMetric.Describe(ch)
    c.lastDomainInfoScrapeDurationSecondsMetric.Describe(ch)
}

func (c *DomainsCollector) reportDomainsMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) {
    c.domainInfoMetric.Reset()

    for _, domain := range objs.Domains {
        internal := "false"
        if domain.Internal.Value {
            internal = "true"
        }

        for _, protocol := range domain.Protocols {
            c.domainInfoMetric.WithLabelValues(
                domain.GUID,
                domain.Name,
                internal,
                protocol,
            ).Set(1)
        }
    }

    c.domainInfoMetric.Collect(ch)
}
