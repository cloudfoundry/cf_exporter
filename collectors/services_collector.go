package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ServicesCollector struct {
	namespace                               string
	cfClient                                *cfclient.Client
	serviceInfoMetric                       *prometheus.GaugeVec
	servicesTotalMetric                     prometheus.Gauge
	lastServicesScrapeErrorMetric           prometheus.Gauge
	lastServicesScrapeTimestampMetric       prometheus.Gauge
	lastServicesScrapeDurationSecondsMetric prometheus.Gauge
}

func NewServicesCollector(namespace string, cfClient *cfclient.Client) *ServicesCollector {
	serviceInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "service",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Service information with a constant '1' value.",
		},
		[]string{"service_id", "service_label"},
	)

	servicesTotalMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "services",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Services.",
		},
	)

	lastServicesScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_services_scrape_error",
			Help:      "Whether the last scrape of Services metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastServicesScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_services_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Services metrics from Cloud Foundry.",
		},
	)

	lastServicesScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_services_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Services metrics from Cloud Foundry.",
		},
	)

	return &ServicesCollector{
		namespace:                               namespace,
		cfClient:                                cfClient,
		serviceInfoMetric:                       serviceInfoMetric,
		servicesTotalMetric:                     servicesTotalMetric,
		lastServicesScrapeErrorMetric:           lastServicesScrapeErrorMetric,
		lastServicesScrapeTimestampMetric:       lastServicesScrapeTimestampMetric,
		lastServicesScrapeDurationSecondsMetric: lastServicesScrapeDurationSecondsMetric,
	}
}

func (c ServicesCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	c.serviceInfoMetric.Reset()

	services, err := c.cfClient.ListServices()
	if err != nil {
		log.Errorf("Error while listing services: %v", err)
		c.reportErrorMetric(true, ch)
		return
	}

	for _, service := range services {
		c.serviceInfoMetric.WithLabelValues(
			service.Guid,
			service.Label,
		).Set(float64(1))
	}

	c.serviceInfoMetric.Collect(ch)

	c.servicesTotalMetric.Set(float64(len(services)))
	c.servicesTotalMetric.Collect(ch)

	c.lastServicesScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastServicesScrapeTimestampMetric.Collect(ch)

	c.lastServicesScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastServicesScrapeDurationSecondsMetric.Collect(ch)

	c.reportErrorMetric(false, ch)
}

func (c ServicesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.serviceInfoMetric.Describe(ch)
	c.servicesTotalMetric.Describe(ch)
	c.lastServicesScrapeErrorMetric.Describe(ch)
	c.lastServicesScrapeTimestampMetric.Describe(ch)
	c.lastServicesScrapeDurationSecondsMetric.Describe(ch)
}

func (c ServicesCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if errorHappend {
		errorMetric = float64(1)
	}

	c.lastServicesScrapeErrorMetric.Set(errorMetric)
	c.lastServicesScrapeErrorMetric.Collect(ch)
}
