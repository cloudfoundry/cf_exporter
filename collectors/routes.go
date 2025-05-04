package collectors

import (
	"time"

	"github.com/cloudfoundry/cf_exporter/v2/models"
	"github.com/prometheus/client_golang/prometheus"
)

type RoutesCollector struct {
	namespace                             string
	environment                           string
	deployment                            string
	routeInfoMetric                       *prometheus.GaugeVec
	routesScrapesTotalMetric              prometheus.Counter
	routesScrapeErrorsTotalMetric         prometheus.Counter
	lastRoutesScrapeErrorMetric           prometheus.Gauge
	lastRoutesScrapeTimestampMetric       prometheus.Gauge
	lastRoutesScrapeDurationSecondsMetric prometheus.Gauge
}

func NewRoutesCollector(
	namespace string,
	environment string,
	deployment string,
) *RoutesCollector {
	routeInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "route",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Route information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"route_id", "route_host", "route_path", "domain_id", "space_id", "service_instance_id"},
	)

	routesScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "routes_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Routes.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	routesScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "routes_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Routes.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastRoutesScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_routes_scrape_error",
			Help:        "Whether the last scrape of Routes metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastRoutesScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_routes_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Routes metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastRoutesScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_routes_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Routes metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &RoutesCollector{
		namespace:                             namespace,
		environment:                           environment,
		deployment:                            deployment,
		routeInfoMetric:                       routeInfoMetric,
		routesScrapesTotalMetric:              routesScrapesTotalMetric,
		routesScrapeErrorsTotalMetric:         routesScrapeErrorsTotalMetric,
		lastRoutesScrapeErrorMetric:           lastRoutesScrapeErrorMetric,
		lastRoutesScrapeTimestampMetric:       lastRoutesScrapeTimestampMetric,
		lastRoutesScrapeDurationSecondsMetric: lastRoutesScrapeDurationSecondsMetric,
	}
}

func (c RoutesCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.Error != nil {
		errorMetric = float64(1)
		c.routesScrapeErrorsTotalMetric.Inc()
	} else {
		c.reportRoutesMetrics(objs, ch)
	}
	c.routesScrapeErrorsTotalMetric.Collect(ch)
	c.routesScrapesTotalMetric.Inc()
	c.routesScrapesTotalMetric.Collect(ch)
	c.lastRoutesScrapeErrorMetric.Set(errorMetric)
	c.lastRoutesScrapeErrorMetric.Collect(ch)
	c.lastRoutesScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastRoutesScrapeTimestampMetric.Collect(ch)
	c.lastRoutesScrapeDurationSecondsMetric.Set(objs.Took)
	c.lastRoutesScrapeDurationSecondsMetric.Collect(ch)
}

func (c RoutesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.routeInfoMetric.Describe(ch)
	c.routesScrapesTotalMetric.Describe(ch)
	c.routesScrapeErrorsTotalMetric.Describe(ch)
	c.lastRoutesScrapeErrorMetric.Describe(ch)
	c.lastRoutesScrapeTimestampMetric.Describe(ch)
	c.lastRoutesScrapeDurationSecondsMetric.Describe(ch)
}

func (c RoutesCollector) reportRoutesMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	c.routeInfoMetric.Reset()

	for _, route := range objs.Routes {
		serviceGUID := ""
		if binding, ok := objs.RoutesBindings[route.GUID]; ok {
			serviceGUID = binding.ServiceInstanceGUID
		}
		c.routeInfoMetric.WithLabelValues(
			route.GUID,
			route.Host,
			route.Path,
			route.DomainGUID,
			route.SpaceGUID,
			serviceGUID,
		).Set(float64(1))
	}

	c.routeInfoMetric.Collect(ch)
}
