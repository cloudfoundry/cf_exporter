package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ServicesCollector struct {
	namespace                             string
	cfClient                              *cfclient.Client
	serviceInfoDesc                       *prometheus.Desc
	servicesTotalDesc                     *prometheus.Desc
	lastServicesScrapeTimestampDesc       *prometheus.Desc
	lastServicesScrapeDurationSecondsDesc *prometheus.Desc
}

func NewServicesCollector(namespace string, cfClient *cfclient.Client) *ServicesCollector {
	serviceInfoDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "service", "info"),
		"Cloud Foundry Service information.",
		[]string{"service_id", "service_label"},
		nil,
	)

	servicesTotalDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "services", "total"),
		"Total number of Cloud Foundry Services.",
		[]string{},
		nil,
	)

	lastServicesScrapeTimestampDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_services_scrape_timestamp"),
		"Number of seconds since 1970 since last scrape of Services metrics from Cloud Foundry.",
		[]string{},
		nil,
	)

	lastServicesScrapeDurationSecondsDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_services_scrape_duration_seconds"),
		"Duration of the last scrape of Services metrics from Cloud Foundry.",
		[]string{},
		nil,
	)

	return &ServicesCollector{
		namespace:                             namespace,
		cfClient:                              cfClient,
		serviceInfoDesc:                       serviceInfoDesc,
		servicesTotalDesc:                     servicesTotalDesc,
		lastServicesScrapeTimestampDesc:       lastServicesScrapeTimestampDesc,
		lastServicesScrapeDurationSecondsDesc: lastServicesScrapeDurationSecondsDesc,
	}
}

func (c ServicesCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	services, err := c.cfClient.ListServices()
	if err != nil {
		log.Errorf("Error while listing services: %v", err)
		return
	}

	for _, service := range services {
		ch <- prometheus.MustNewConstMetric(
			c.serviceInfoDesc,
			prometheus.GaugeValue,
			float64(1),
			service.Guid,
			service.Label,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.servicesTotalDesc,
		prometheus.GaugeValue,
		float64(len(services)),
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastServicesScrapeTimestampDesc,
		prometheus.GaugeValue,
		float64(time.Now().Unix()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.lastServicesScrapeDurationSecondsDesc,
		prometheus.GaugeValue,
		time.Since(begun).Seconds(),
	)
}

func (c ServicesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.serviceInfoDesc
	ch <- c.servicesTotalDesc
	ch <- c.lastServicesScrapeTimestampDesc
	ch <- c.lastServicesScrapeDurationSecondsDesc
}
