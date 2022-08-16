package collectors

import (
	"time"

	"github.com/bosh-prometheus/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type ServicesCollector struct {
	namespace                               string
	environment                             string
	deployment                              string
	serviceInfoMetric                       *prometheus.GaugeVec
	servicesScrapesTotalMetric              prometheus.Counter
	servicesScrapeErrorsTotalMetric         prometheus.Counter
	lastServicesScrapeErrorMetric           prometheus.Gauge
	lastServicesScrapeTimestampMetric       prometheus.Gauge
	lastServicesScrapeDurationSecondsMetric prometheus.Gauge
}

func NewServicesCollector(
	namespace string,
	environment string,
	deployment string,
) *ServicesCollector {
	serviceInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "service",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Service information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"service_id", "service_label"},
	)

	servicesScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "services_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Services.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	servicesScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "services_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Services.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServicesScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_services_scrape_error",
			Help:        "Whether the last scrape of Services metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServicesScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_services_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Services metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServicesScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_services_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Services metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &ServicesCollector{
		namespace:                               namespace,
		environment:                             environment,
		deployment:                              deployment,
		serviceInfoMetric:                       serviceInfoMetric,
		servicesScrapesTotalMetric:              servicesScrapesTotalMetric,
		servicesScrapeErrorsTotalMetric:         servicesScrapeErrorsTotalMetric,
		lastServicesScrapeErrorMetric:           lastServicesScrapeErrorMetric,
		lastServicesScrapeTimestampMetric:       lastServicesScrapeTimestampMetric,
		lastServicesScrapeDurationSecondsMetric: lastServicesScrapeDurationSecondsMetric,
	}
}

func (c ServicesCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	err := objs.Error
	if objs.Error != nil {
		errorMetric = float64(1)
		c.servicesScrapeErrorsTotalMetric.Inc()
	} else {
		err = c.reportServicesMetrics(objs, ch)
		if err != nil {
			log.Error(err)
			errorMetric = float64(1)
			c.servicesScrapeErrorsTotalMetric.Inc()
		}
	}

	c.servicesScrapeErrorsTotalMetric.Collect(ch)
	c.servicesScrapesTotalMetric.Inc()
	c.servicesScrapesTotalMetric.Collect(ch)
	c.lastServicesScrapeErrorMetric.Set(errorMetric)
	c.lastServicesScrapeErrorMetric.Collect(ch)
	c.lastServicesScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastServicesScrapeTimestampMetric.Collect(ch)
	c.lastServicesScrapeDurationSecondsMetric.Set(objs.Took)
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

func (c ServicesCollector) reportServicesMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) error {
	c.serviceInfoMetric.Reset()

	for _, cService := range objs.ServiceOfferings {
		c.serviceInfoMetric.WithLabelValues(
			cService.GUID,
			cService.Name,
		).Set(float64(1))
	}

	c.serviceInfoMetric.Collect(ch)
	return nil
}
