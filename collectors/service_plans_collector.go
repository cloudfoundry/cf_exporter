package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ServicePlansCollector struct {
	namespace                                   string
	environment                                 string
	deployment                                  string
	cfClient                                    *cfclient.Client
	servicePlanInfoMetric                       *prometheus.GaugeVec
	servicePlansScrapesTotalMetric              prometheus.Counter
	servicePlansScrapeErrorsTotalMetric         prometheus.Counter
	lastServicePlansScrapeErrorMetric           prometheus.Gauge
	lastServicePlansScrapeTimestampMetric       prometheus.Gauge
	lastServicePlansScrapeDurationSecondsMetric prometheus.Gauge
}

func NewServicePlansCollector(
	namespace string,
	environment string,
	deployment string,
	cfClient *cfclient.Client,
) *ServicePlansCollector {
	servicePlanInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "service_plan",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Service Plan information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"service_plan_id", "service_plan_name", "service_id"},
	)

	servicePlansScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "service_plans_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Service Plans.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	servicePlansScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "service_plans_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Service Plans.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServicePlansScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_plans_scrape_error",
			Help:        "Whether the last scrape of Service Plans metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServicePlansScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_plans_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Service Plans metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastServicePlansScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_service_plans_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Service Plans metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &ServicePlansCollector{
		namespace:                                   namespace,
		environment:                                 environment,
		deployment:                                  deployment,
		cfClient:                                    cfClient,
		servicePlanInfoMetric:                       servicePlanInfoMetric,
		servicePlansScrapesTotalMetric:              servicePlansScrapesTotalMetric,
		servicePlansScrapeErrorsTotalMetric:         servicePlansScrapeErrorsTotalMetric,
		lastServicePlansScrapeErrorMetric:           lastServicePlansScrapeErrorMetric,
		lastServicePlansScrapeTimestampMetric:       lastServicePlansScrapeTimestampMetric,
		lastServicePlansScrapeDurationSecondsMetric: lastServicePlansScrapeDurationSecondsMetric,
	}
}

func (c ServicePlansCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportServicePlansMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.servicePlansScrapeErrorsTotalMetric.Inc()
	}
	c.servicePlansScrapeErrorsTotalMetric.Collect(ch)

	c.servicePlansScrapesTotalMetric.Inc()
	c.servicePlansScrapesTotalMetric.Collect(ch)

	c.lastServicePlansScrapeErrorMetric.Set(errorMetric)
	c.lastServicePlansScrapeErrorMetric.Collect(ch)

	c.lastServicePlansScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastServicePlansScrapeTimestampMetric.Collect(ch)

	c.lastServicePlansScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastServicePlansScrapeDurationSecondsMetric.Collect(ch)
}

func (c ServicePlansCollector) Describe(ch chan<- *prometheus.Desc) {
	c.servicePlanInfoMetric.Describe(ch)
	c.servicePlansScrapesTotalMetric.Describe(ch)
	c.servicePlansScrapeErrorsTotalMetric.Describe(ch)
	c.lastServicePlansScrapeErrorMetric.Describe(ch)
	c.lastServicePlansScrapeTimestampMetric.Describe(ch)
	c.lastServicePlansScrapeDurationSecondsMetric.Describe(ch)
}

func (c ServicePlansCollector) reportServicePlansMetrics(ch chan<- prometheus.Metric) error {
	c.servicePlanInfoMetric.Reset()

	servicePlans, err := c.cfClient.ListServicePlans()
	if err != nil {
		log.Errorf("Error while listing service planss: %v", err)
		return err
	}

	for _, servicePlan := range servicePlans {
		c.servicePlanInfoMetric.WithLabelValues(
			servicePlan.Guid,
			servicePlan.Name,
			servicePlan.ServiceGuid,
		).Set(float64(1))
	}

	c.servicePlanInfoMetric.Collect(ch)

	return nil
}
