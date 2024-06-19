package collectors

import (
	"time"

	"github.com/cloudfoundry/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
)

type RouteBindingsCollector struct {
	namespace                                           string
	environment                                         string
	deployment                                          string
	serviceRouteBindingInfoMetric                       *prometheus.GaugeVec
	serviceRouteBindingsScrapesTotalMetric              prometheus.Counter
	serviceRouteBindingsScrapeErrorsTotalMetric         prometheus.Counter
	lastServiceRouteBindingsScrapeErrorMetric           prometheus.Gauge
	lastServiceRouteBindingsScrapeTimestampMetric       prometheus.Gauge
	lastServiceRouteBindingsScrapeDurationSecondsMetric prometheus.Gauge
}

func NewRouteBindingsCollector(
	namespace string,
	environment string,
	deployment string,
) *RouteBindingsCollector {
	serviceRouteBindingInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "service_route_bindings",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Service Route Binding information with a constant '1' value",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"service_route_binding_id", "route_service_url", "service_instance_id", "route_id"},
	)

	serviceRouteBindingsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "service_route_bindings_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Service Route Bindings.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	serviceRouteBindingsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "service_route_bindings_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Service Route Bindings.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServiceRouteBindingsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_route_bindings_scrape_error",
			Help:        "Whether the last scrape of Service Route Bindings metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServiceRouteBindingsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_route_bindings_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Service Route Bindings metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServiceRouteBindingsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_route_bindings_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Service Route Bindings metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &RouteBindingsCollector{
		namespace:                              namespace,
		environment:                            environment,
		deployment:                             deployment,
		serviceRouteBindingInfoMetric:          serviceRouteBindingInfoMetric,
		serviceRouteBindingsScrapesTotalMetric: serviceRouteBindingsScrapesTotalMetric,
		serviceRouteBindingsScrapeErrorsTotalMetric:         serviceRouteBindingsScrapeErrorsTotalMetric,
		lastServiceRouteBindingsScrapeErrorMetric:           lastServiceRouteBindingsScrapeErrorMetric,
		lastServiceRouteBindingsScrapeTimestampMetric:       lastServiceRouteBindingsScrapeTimestampMetric,
		lastServiceRouteBindingsScrapeDurationSecondsMetric: lastServiceRouteBindingsScrapeDurationSecondsMetric,
	}
}

func (c RouteBindingsCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.Error != nil {
		errorMetric = float64(1)
		c.serviceRouteBindingsScrapeErrorsTotalMetric.Inc()
	} else {
		c.reportRouteBindingsMetrics(objs, ch)
	}
	c.serviceRouteBindingsScrapeErrorsTotalMetric.Collect(ch)
	c.serviceRouteBindingsScrapesTotalMetric.Inc()
	c.serviceRouteBindingsScrapesTotalMetric.Collect(ch)
	c.lastServiceRouteBindingsScrapeErrorMetric.Set(errorMetric)
	c.lastServiceRouteBindingsScrapeErrorMetric.Collect(ch)
	c.lastServiceRouteBindingsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastServiceRouteBindingsScrapeTimestampMetric.Collect(ch)
	c.lastServiceRouteBindingsScrapeDurationSecondsMetric.Set(objs.Took)
	c.lastServiceRouteBindingsScrapeDurationSecondsMetric.Collect(ch)
}

func (c RouteBindingsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.serviceRouteBindingInfoMetric.Describe(ch)
	c.serviceRouteBindingsScrapesTotalMetric.Describe(ch)
	c.serviceRouteBindingsScrapeErrorsTotalMetric.Describe(ch)
	c.lastServiceRouteBindingsScrapeErrorMetric.Describe(ch)
	c.lastServiceRouteBindingsScrapeTimestampMetric.Describe(ch)
	c.lastServiceRouteBindingsScrapeDurationSecondsMetric.Describe(ch)
}

func (c RouteBindingsCollector) reportRouteBindingsMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	c.serviceRouteBindingInfoMetric.Reset()
	for _, cItem := range objs.ServiceRouteBindings {
		c.serviceRouteBindingInfoMetric.WithLabelValues(
			cItem.GUID,
			cItem.RouteServiceURL,
			cItem.ServiceInstanceGUID,
			cItem.RouteGUID,
		).Set(float64(1))
	}
	c.serviceRouteBindingInfoMetric.Collect(ch)
}
