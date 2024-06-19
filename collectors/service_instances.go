package collectors

import (
	"time"

	"code.cloudfoundry.org/cli/resources"
	"github.com/cloudfoundry/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
)

type ServiceInstancesCollector struct {
	namespace                                       string
	environment                                     string
	deployment                                      string
	serviceInstanceInfoMetric                       *prometheus.GaugeVec
	serviceInstancesScrapesTotalMetric              prometheus.Counter
	serviceInstancesScrapeErrorsTotalMetric         prometheus.Counter
	lastServiceInstancesScrapeErrorMetric           prometheus.Gauge
	lastServiceInstancesScrapeTimestampMetric       prometheus.Gauge
	lastServiceInstancesScrapeDurationSecondsMetric prometheus.Gauge
}

func NewServiceInstancesCollector(
	namespace string,
	environment string,
	deployment string,
) *ServiceInstancesCollector {
	serviceInstanceInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "service_instance",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Service Instance information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"service_instance_id", "service_instance_name", "service_plan_id", "space_id", "type", "last_operation_type", "last_operation_state"},
	)

	serviceInstancesScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "service_instances_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Service Instances.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	serviceInstancesScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "service_instances_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Service Instances.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServiceInstancesScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_instances_scrape_error",
			Help:        "Whether the last scrape of Service Instances metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServiceInstancesScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_instances_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Service Instances metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServiceInstancesScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_instances_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Service Instances metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &ServiceInstancesCollector{
		namespace:                                       namespace,
		environment:                                     environment,
		deployment:                                      deployment,
		serviceInstanceInfoMetric:                       serviceInstanceInfoMetric,
		serviceInstancesScrapesTotalMetric:              serviceInstancesScrapesTotalMetric,
		serviceInstancesScrapeErrorsTotalMetric:         serviceInstancesScrapeErrorsTotalMetric,
		lastServiceInstancesScrapeErrorMetric:           lastServiceInstancesScrapeErrorMetric,
		lastServiceInstancesScrapeTimestampMetric:       lastServiceInstancesScrapeTimestampMetric,
		lastServiceInstancesScrapeDurationSecondsMetric: lastServiceInstancesScrapeDurationSecondsMetric,
	}
}

func (c ServiceInstancesCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.Error != nil {
		errorMetric = float64(1)
		c.serviceInstancesScrapeErrorsTotalMetric.Inc()
	} else {
		c.reportServiceInstancesMetrics(objs, ch)
	}

	c.serviceInstancesScrapeErrorsTotalMetric.Collect(ch)
	c.serviceInstancesScrapesTotalMetric.Inc()
	c.serviceInstancesScrapesTotalMetric.Collect(ch)
	c.lastServiceInstancesScrapeErrorMetric.Set(errorMetric)
	c.lastServiceInstancesScrapeErrorMetric.Collect(ch)
	c.lastServiceInstancesScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastServiceInstancesScrapeTimestampMetric.Collect(ch)
	c.lastServiceInstancesScrapeDurationSecondsMetric.Set(objs.Took)
	c.lastServiceInstancesScrapeDurationSecondsMetric.Collect(ch)
}

func (c ServiceInstancesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.serviceInstanceInfoMetric.Describe(ch)
	c.serviceInstancesScrapesTotalMetric.Describe(ch)
	c.serviceInstancesScrapeErrorsTotalMetric.Describe(ch)
	c.lastServiceInstancesScrapeErrorMetric.Describe(ch)
	c.lastServiceInstancesScrapeTimestampMetric.Describe(ch)
	c.lastServiceInstancesScrapeDurationSecondsMetric.Describe(ch)
}

// reportServiceInstancesMetrics
// 1. v0 compatibility
func (c ServiceInstancesCollector) reportServiceInstancesMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	c.serviceInstanceInfoMetric.Reset()

	for _, cElem := range objs.ServiceInstances {
		// 1.
		sType := string(cElem.Type)
		if cElem.Type == resources.ManagedServiceInstance {
			sType = "managed_service_instance"
		}

		c.serviceInstanceInfoMetric.WithLabelValues(
			cElem.GUID,
			cElem.Name,
			cElem.ServicePlanGUID,
			cElem.SpaceGUID,
			sType,
			string(cElem.LastOperation.Type),
			string(cElem.LastOperation.State),
		).Set(float64(1))
	}

	c.serviceInstanceInfoMetric.Collect(ch)
}
