package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ServiceInstancesCollector struct {
	namespace                                       string
	deploymentName                                  string
	cfClient                                        *cfclient.Client
	serviceInstanceInfoMetric                       *prometheus.GaugeVec
	serviceInstancesScrapesTotalMetric              *prometheus.CounterVec
	serviceInstancesScrapeErrorsTotalMetric         *prometheus.CounterVec
	lastServiceInstancesScrapeErrorMetric           *prometheus.GaugeVec
	lastServiceInstancesScrapeTimestampMetric       *prometheus.GaugeVec
	lastServiceInstancesScrapeDurationSecondsMetric *prometheus.GaugeVec
}

func NewServiceInstancesCollector(namespace string, deploymentName string, cfClient *cfclient.Client) *ServiceInstancesCollector {
	serviceInstanceInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "service_instance",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Service Instance information with a constant '1' value.",
		},
		[]string{"deployment", "service_instance_id", "service_instance_name", "service_plan_id", "space_id", "type"},
	)

	serviceInstancesScrapesTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "service_instances_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Cloud Foundry Service Instances.",
		},
		[]string{"deployment"},
	)

	serviceInstancesScrapeErrorsTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "service_instances_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrape error of Cloud Foundry Service Instances.",
		},
		[]string{"deployment"},
	)

	lastServiceInstancesScrapeErrorMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_service_instances_scrape_error",
			Help:      "Whether the last scrape of Service Instances metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
		[]string{"deployment"},
	)

	lastServiceInstancesScrapeTimestampMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_service_instances_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Service Instances metrics from Cloud Foundry.",
		},
		[]string{"deployment"},
	)

	lastServiceInstancesScrapeDurationSecondsMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_service_instances_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Service Instances metrics from Cloud Foundry.",
		},
		[]string{"deployment"},
	)

	return &ServiceInstancesCollector{
		namespace:                                       namespace,
		deploymentName:                                  deploymentName,
		cfClient:                                        cfClient,
		serviceInstanceInfoMetric:                       serviceInstanceInfoMetric,
		serviceInstancesScrapesTotalMetric:              serviceInstancesScrapesTotalMetric,
		serviceInstancesScrapeErrorsTotalMetric:         serviceInstancesScrapeErrorsTotalMetric,
		lastServiceInstancesScrapeErrorMetric:           lastServiceInstancesScrapeErrorMetric,
		lastServiceInstancesScrapeTimestampMetric:       lastServiceInstancesScrapeTimestampMetric,
		lastServiceInstancesScrapeDurationSecondsMetric: lastServiceInstancesScrapeDurationSecondsMetric,
	}
}

func (c ServiceInstancesCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportServiceInstancesMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.serviceInstancesScrapeErrorsTotalMetric.WithLabelValues(c.deploymentName).Inc()
	}
	c.serviceInstancesScrapeErrorsTotalMetric.Collect(ch)

	c.serviceInstancesScrapesTotalMetric.WithLabelValues(c.deploymentName).Inc()
	c.serviceInstancesScrapesTotalMetric.Collect(ch)

	c.lastServiceInstancesScrapeErrorMetric.WithLabelValues(c.deploymentName).Set(errorMetric)
	c.lastServiceInstancesScrapeErrorMetric.Collect(ch)

	c.lastServiceInstancesScrapeTimestampMetric.WithLabelValues(c.deploymentName).Set(float64(time.Now().Unix()))
	c.lastServiceInstancesScrapeTimestampMetric.Collect(ch)

	c.lastServiceInstancesScrapeDurationSecondsMetric.WithLabelValues(c.deploymentName).Set(time.Since(begun).Seconds())
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

func (c ServiceInstancesCollector) reportServiceInstancesMetrics(ch chan<- prometheus.Metric) error {
	c.serviceInstanceInfoMetric.Reset()

	serviceInstances, err := c.cfClient.ListServiceInstances()
	if err != nil {
		log.Errorf("Error while listing service instances: %v", err)
		return err
	}

	for _, serviceInstance := range serviceInstances {
		c.serviceInstanceInfoMetric.WithLabelValues(
			c.deploymentName,
			serviceInstance.Guid,
			serviceInstance.Name,
			serviceInstance.ServicePlanGuid,
			serviceInstance.SpaceGuid,
			serviceInstance.Type,
		).Set(float64(1))
	}

	c.serviceInstanceInfoMetric.Collect(ch)

	return nil
}
