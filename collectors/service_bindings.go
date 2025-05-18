package collectors

import (
	"time"

	"github.com/cloudfoundry/cf_exporter/v2/models"
	"github.com/prometheus/client_golang/prometheus"
)

type ServiceBindingsCollector struct {
	namespace                                      string
	environment                                    string
	deployment                                     string
	serviceBindingInfoMetric                       *prometheus.GaugeVec
	serviceBindingsScrapesTotalMetric              prometheus.Counter
	serviceBindingsScrapeErrorsTotalMetric         prometheus.Counter
	lastServiceBindingsScrapeErrorMetric           prometheus.Gauge
	lastServiceBindingsScrapeTimestampMetric       prometheus.Gauge
	lastServiceBindingsScrapeDurationSecondsMetric prometheus.Gauge
}

func NewServiceBindingsCollector(
	namespace string,
	environment string,
	deployment string,
) *ServiceBindingsCollector {
	serviceBindingInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "service_binding",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Service Binding information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"service_binding_id", "application_id", "service_instance_id"},
	)

	serviceBindingsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "service_bindings_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Service Bindings.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	serviceBindingsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "service_bindings_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Service Bindings.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServiceBindingsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_bindings_scrape_error",
			Help:        "Whether the last scrape of Service Bindings metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServiceBindingsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_bindings_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Service Bindings metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServiceBindingsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_bindings_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Service Bindings metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &ServiceBindingsCollector{
		namespace:                                      namespace,
		environment:                                    environment,
		deployment:                                     deployment,
		serviceBindingInfoMetric:                       serviceBindingInfoMetric,
		serviceBindingsScrapesTotalMetric:              serviceBindingsScrapesTotalMetric,
		serviceBindingsScrapeErrorsTotalMetric:         serviceBindingsScrapeErrorsTotalMetric,
		lastServiceBindingsScrapeErrorMetric:           lastServiceBindingsScrapeErrorMetric,
		lastServiceBindingsScrapeTimestampMetric:       lastServiceBindingsScrapeTimestampMetric,
		lastServiceBindingsScrapeDurationSecondsMetric: lastServiceBindingsScrapeDurationSecondsMetric,
	}
}

func (c ServiceBindingsCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.Error != nil {
		errorMetric = float64(1)
		c.serviceBindingsScrapeErrorsTotalMetric.Inc()
	} else {
		c.reportServiceBindingsMetrics(objs, ch)
	}
	c.serviceBindingsScrapeErrorsTotalMetric.Collect(ch)
	c.serviceBindingsScrapesTotalMetric.Inc()
	c.serviceBindingsScrapesTotalMetric.Collect(ch)
	c.lastServiceBindingsScrapeErrorMetric.Set(errorMetric)
	c.lastServiceBindingsScrapeErrorMetric.Collect(ch)
	c.lastServiceBindingsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastServiceBindingsScrapeTimestampMetric.Collect(ch)
	c.lastServiceBindingsScrapeDurationSecondsMetric.Set(objs.Took)
	c.lastServiceBindingsScrapeDurationSecondsMetric.Collect(ch)
}

func (c ServiceBindingsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.serviceBindingInfoMetric.Describe(ch)
	c.serviceBindingsScrapesTotalMetric.Describe(ch)
	c.serviceBindingsScrapeErrorsTotalMetric.Describe(ch)
	c.lastServiceBindingsScrapeErrorMetric.Describe(ch)
	c.lastServiceBindingsScrapeTimestampMetric.Describe(ch)
	c.lastServiceBindingsScrapeDurationSecondsMetric.Describe(ch)
}

func (c ServiceBindingsCollector) reportServiceBindingsMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	c.serviceBindingInfoMetric.Reset()

	for _, cItem := range objs.ServiceBindings {
		c.serviceBindingInfoMetric.WithLabelValues(
			cItem.GUID,
			cItem.AppGUID,
			cItem.ServiceInstanceGUID,
		).Set(float64(1))
	}
	c.serviceBindingInfoMetric.Collect(ch)
}
