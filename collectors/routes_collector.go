package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type RoutesCollector struct {
	namespace                             string
	environment                           string
	deploymentName                        string
	cfClient                              *cfclient.Client
	routeInfoMetric                       *prometheus.GaugeVec
	routesScrapesTotalMetric              *prometheus.CounterVec
	routesScrapeErrorsTotalMetric         *prometheus.CounterVec
	lastRoutesScrapeErrorMetric           *prometheus.GaugeVec
	lastRoutesScrapeTimestampMetric       *prometheus.GaugeVec
	lastRoutesScrapeDurationSecondsMetric *prometheus.GaugeVec
}

func NewRoutesCollector(
	namespace string,
	environment string,
	deploymentName string,
	cfClient *cfclient.Client,
) *RoutesCollector {
	routeInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "route",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Route information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment", "route_id", "route_host", "route_path", "domain_id", "space_id", "service_instance_id"},
	)

	routesScrapesTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "routes_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Routes.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	routesScrapeErrorsTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "routes_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Routes.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	lastRoutesScrapeErrorMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_routes_scrape_error",
			Help:        "Whether the last scrape of Routes metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	lastRoutesScrapeTimestampMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_routes_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Routes metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	lastRoutesScrapeDurationSecondsMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_routes_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Routes metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment},
		},
		[]string{"deployment"},
	)

	return &RoutesCollector{
		namespace:                             namespace,
		environment:                           environment,
		deploymentName:                        deploymentName,
		cfClient:                              cfClient,
		routeInfoMetric:                       routeInfoMetric,
		routesScrapesTotalMetric:              routesScrapesTotalMetric,
		routesScrapeErrorsTotalMetric:         routesScrapeErrorsTotalMetric,
		lastRoutesScrapeErrorMetric:           lastRoutesScrapeErrorMetric,
		lastRoutesScrapeTimestampMetric:       lastRoutesScrapeTimestampMetric,
		lastRoutesScrapeDurationSecondsMetric: lastRoutesScrapeDurationSecondsMetric,
	}
}

func (c RoutesCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportRoutesMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.routesScrapeErrorsTotalMetric.WithLabelValues(c.deploymentName).Inc()
	}
	c.routesScrapeErrorsTotalMetric.Collect(ch)

	c.routesScrapesTotalMetric.WithLabelValues(c.deploymentName).Inc()
	c.routesScrapesTotalMetric.Collect(ch)

	c.lastRoutesScrapeErrorMetric.WithLabelValues(c.deploymentName).Set(errorMetric)
	c.lastRoutesScrapeErrorMetric.Collect(ch)

	c.lastRoutesScrapeTimestampMetric.WithLabelValues(c.deploymentName).Set(float64(time.Now().Unix()))
	c.lastRoutesScrapeTimestampMetric.Collect(ch)

	c.lastRoutesScrapeDurationSecondsMetric.WithLabelValues(c.deploymentName).Set(time.Since(begun).Seconds())
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

func (c RoutesCollector) reportRoutesMetrics(ch chan<- prometheus.Metric) error {
	c.routeInfoMetric.Reset()

	routes, err := c.cfClient.ListRoutes()
	if err != nil {
		log.Errorf("Error while listing routes: %v", err)
		return err
	}

	for _, route := range routes {
		c.routeInfoMetric.WithLabelValues(
			c.deploymentName,
			route.Guid,
			route.Host,
			route.Path,
			route.DomainGuid,
			route.SpaceGuid,
			route.ServiceInstanceGuid,
		).Set(float64(1))
	}

	c.routeInfoMetric.Collect(ch)

	return nil
}
