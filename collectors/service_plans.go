package collectors

import (
	"time"

	"github.com/bosh-prometheus/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
)

type ServicePlansCollector struct {
	namespace                                   string
	environment                                 string
	deployment                                  string
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
		servicePlanInfoMetric:                       servicePlanInfoMetric,
		servicePlansScrapesTotalMetric:              servicePlansScrapesTotalMetric,
		servicePlansScrapeErrorsTotalMetric:         servicePlansScrapeErrorsTotalMetric,
		lastServicePlansScrapeErrorMetric:           lastServicePlansScrapeErrorMetric,
		lastServicePlansScrapeTimestampMetric:       lastServicePlansScrapeTimestampMetric,
		lastServicePlansScrapeDurationSecondsMetric: lastServicePlansScrapeDurationSecondsMetric,
	}
}

func (c ServicePlansCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.Error != nil {
		errorMetric = float64(1)
		c.servicePlansScrapeErrorsTotalMetric.Inc()
	} else {
		c.reportServicePlansMetrics(objs, ch)
	}

	c.servicePlansScrapeErrorsTotalMetric.Collect(ch)
	c.servicePlansScrapesTotalMetric.Inc()
	c.servicePlansScrapesTotalMetric.Collect(ch)
	c.lastServicePlansScrapeErrorMetric.Set(errorMetric)
	c.lastServicePlansScrapeErrorMetric.Collect(ch)
	c.lastServicePlansScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastServicePlansScrapeTimestampMetric.Collect(ch)
	c.lastServicePlansScrapeDurationSecondsMetric.Set(objs.Took)
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

func (c ServicePlansCollector) reportServicePlansMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	c.servicePlanInfoMetric.Reset()

	for _, cElem := range objs.ServicePlans {
		c.servicePlanInfoMetric.WithLabelValues(
			cElem.GUID,
			cElem.Name,
			cElem.ServiceOfferingGUID,
		).Set(float64(1))
	}

	c.servicePlanInfoMetric.Collect(ch)
}
