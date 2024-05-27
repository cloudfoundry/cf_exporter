package collectors

import (
    "github.com/bosh-prometheus/cf_exporter/models"
    "github.com/prometheus/client_golang/prometheus"
)

type DomainsCollector struct {
    namespace            string
    environment          string
    deployment           string
    domainInfoMetric     *prometheus.GaugeVec
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

    return &DomainsCollector{
        namespace:         namespace,
        environment:       environment,
        deployment:        deployment,
        domainInfoMetric:  domainInfoMetric,
    }
}

func (c *DomainsCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
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

func (c *DomainsCollector) Describe(ch chan<- *prometheus.Desc) {
    c.domainInfoMetric.Describe(ch)
}
