package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ServicesCollector struct {
	namespace                         string
	cfClient                          *cfclient.Client
	serviceInfo                       *prometheus.GaugeVec
	servicesTotal                     prometheus.Gauge
	lastServicesScrapeError           prometheus.Gauge
	lastServicesScrapeTimestamp       prometheus.Gauge
	lastServicesScrapeDurationSeconds prometheus.Gauge
}

func NewServicesCollector(namespace string, cfClient *cfclient.Client) *ServicesCollector {
	serviceInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "service",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Service information with a constant '1' value.",
		},
		[]string{"service_id", "service_label"},
	)

	servicesTotal := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "services",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Services.",
		},
	)

	lastServicesScrapeError := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_services_scrape_error",
			Help:      "Whether the last scrape of Services metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastServicesScrapeTimestamp := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_services_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Services metrics from Cloud Foundry.",
		},
	)

	lastServicesScrapeDurationSeconds := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_services_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Services metrics from Cloud Foundry.",
		},
	)

	return &ServicesCollector{
		namespace:                         namespace,
		cfClient:                          cfClient,
		serviceInfo:                       serviceInfo,
		servicesTotal:                     servicesTotal,
		lastServicesScrapeError:           lastServicesScrapeError,
		lastServicesScrapeTimestamp:       lastServicesScrapeTimestamp,
		lastServicesScrapeDurationSeconds: lastServicesScrapeDurationSeconds,
	}
}

func (c ServicesCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	services, err := c.cfClient.ListServices()
	if err != nil {
		log.Errorf("Error while listing services: %v", err)
		c.reportErrorMetric(true, ch)
		return
	}

	for _, service := range services {
		c.serviceInfo.WithLabelValues(
			service.Guid,
			service.Label,
		).Set(float64(1))
	}
	c.serviceInfo.Collect(ch)

	c.servicesTotal.Set(float64(len(services)))
	c.servicesTotal.Collect(ch)

	c.lastServicesScrapeTimestamp.Set(float64(time.Now().Unix()))
	c.lastServicesScrapeTimestamp.Collect(ch)

	c.lastServicesScrapeDurationSeconds.Set(time.Since(begun).Seconds())
	c.lastServicesScrapeDurationSeconds.Collect(ch)

	c.reportErrorMetric(false, ch)
}

func (c ServicesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.serviceInfo.Describe(ch)
	c.servicesTotal.Describe(ch)
	c.lastServicesScrapeError.Describe(ch)
	c.lastServicesScrapeTimestamp.Describe(ch)
	c.lastServicesScrapeDurationSeconds.Describe(ch)
}

func (c ServicesCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if errorHappend {
		errorMetric = float64(1)
	}

	c.lastServicesScrapeError.Set(errorMetric)
	c.lastServicesScrapeError.Collect(ch)

}
