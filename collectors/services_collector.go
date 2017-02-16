package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ServicesCollector struct {
	namespace                               string
	environment                             string
	deploymentName                          string
	cfClient                                *cfclient.Client
	serviceInfoMetric                       *prometheus.GaugeVec
	servicesScrapesTotalMetric              *prometheus.CounterVec
	servicesScrapeErrorsTotalMetric         *prometheus.CounterVec
	lastServicesScrapeErrorMetric           *prometheus.GaugeVec
	lastServicesScrapeTimestampMetric       *prometheus.GaugeVec
	lastServicesScrapeDurationSecondsMetric *prometheus.GaugeVec
}

func NewServicesCollector(
	namespace string,
	environment string,
	deploymentName string,
	cfClient *cfclient.Client,
) *ServicesCollector {
	serviceInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "service",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Service information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment", "service_id", "service_label"},
	)

	servicesScrapesTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "services_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Services.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	servicesScrapeErrorsTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "services_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Services.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	lastServicesScrapeErrorMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_services_scrape_error",
			Help:        "Whether the last scrape of Services metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	lastServicesScrapeTimestampMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_services_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Services metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	lastServicesScrapeDurationSecondsMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_services_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Services metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	return &ServicesCollector{
		namespace:                               namespace,
		environment:                             environment,
		deploymentName:                          deploymentName,
		cfClient:                                cfClient,
		serviceInfoMetric:                       serviceInfoMetric,
		servicesScrapesTotalMetric:              servicesScrapesTotalMetric,
		servicesScrapeErrorsTotalMetric:         servicesScrapeErrorsTotalMetric,
		lastServicesScrapeErrorMetric:           lastServicesScrapeErrorMetric,
		lastServicesScrapeTimestampMetric:       lastServicesScrapeTimestampMetric,
		lastServicesScrapeDurationSecondsMetric: lastServicesScrapeDurationSecondsMetric,
	}
}

func (c ServicesCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportServicesMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.servicesScrapeErrorsTotalMetric.WithLabelValues(c.deploymentName).Inc()
	}
	c.servicesScrapeErrorsTotalMetric.Collect(ch)

	c.servicesScrapesTotalMetric.WithLabelValues(c.deploymentName).Inc()
	c.servicesScrapesTotalMetric.Collect(ch)

	c.lastServicesScrapeErrorMetric.WithLabelValues(c.deploymentName).Set(errorMetric)
	c.lastServicesScrapeErrorMetric.Collect(ch)

	c.lastServicesScrapeTimestampMetric.WithLabelValues(c.deploymentName).Set(float64(time.Now().Unix()))
	c.lastServicesScrapeTimestampMetric.Collect(ch)

	c.lastServicesScrapeDurationSecondsMetric.WithLabelValues(c.deploymentName).Set(time.Since(begun).Seconds())
	c.lastServicesScrapeDurationSecondsMetric.Collect(ch)
}

func (c ServicesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.serviceInfoMetric.Describe(ch)
	c.servicesScrapesTotalMetric.Describe(ch)
	c.servicesScrapeErrorsTotalMetric.Describe(ch)
	c.lastServicesScrapeErrorMetric.Describe(ch)
	c.lastServicesScrapeTimestampMetric.Describe(ch)
	c.lastServicesScrapeDurationSecondsMetric.Describe(ch)
}

func (c ServicesCollector) reportServicesMetrics(ch chan<- prometheus.Metric) error {
	c.serviceInfoMetric.Reset()

	services, err := c.cfClient.ListServices()
	if err != nil {
		log.Errorf("Error while listing services: %v", err)
		return err
	}

	for _, service := range services {
		c.serviceInfoMetric.WithLabelValues(
			c.deploymentName,
			service.Guid,
			service.Label,
		).Set(float64(1))
	}

	c.serviceInfoMetric.Collect(ch)

	return nil
}
