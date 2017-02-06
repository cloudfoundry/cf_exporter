package collectors

import (
	"sync"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ApplicationEventsCollector struct {
	namespace                                        string
	deploymentName                                   string
	cfClient                                         *cfclient.Client
	applicationEventsTotalMetric                     *prometheus.CounterVec
	applicationEventsScrapesTotalMetric              *prometheus.CounterVec
	applicationEventsScrapeErrorsTotalMetric         *prometheus.CounterVec
	lastApplicationEventsScrapeErrorMetric           *prometheus.GaugeVec
	lastApplicationEventsScrapeTimestampMetric       *prometheus.GaugeVec
	lastApplicationEventsScrapeDurationSecondsMetric *prometheus.GaugeVec
}

func NewApplicationEventsCollector(namespace string, deploymentName string, cfClient *cfclient.Client) *ApplicationEventsCollector {
	applicationEventsTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "application_events",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Application Events.",
		},
		[]string{"deployment", "application_id", "event_type"},
	)

	applicationEventsScrapesTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "application_events_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Cloud Foundry Application Events.",
		},
		[]string{"deployment"},
	)

	applicationEventsScrapeErrorsTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "application_events_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrape errors of Cloud Foundry Application Events.",
		},
		[]string{"deployment"},
	)

	lastApplicationEventsScrapeErrorMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_application_events_scrape_error",
			Help:      "Whether the last scrape of Application Events metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
		[]string{"deployment"},
	)

	lastApplicationEventsScrapeTimestampMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_application_events_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Application Events metrics from Cloud Foundry.",
		},
		[]string{"deployment"},
	)

	lastApplicationEventsScrapeDurationSecondsMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_application_events_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Application Events metrics from Cloud Foundry.",
		},
		[]string{"deployment"},
	)

	return &ApplicationEventsCollector{
		namespace:                                        namespace,
		deploymentName:                                   deploymentName,
		cfClient:                                         cfClient,
		applicationEventsTotalMetric:                     applicationEventsTotalMetric,
		applicationEventsScrapesTotalMetric:              applicationEventsScrapesTotalMetric,
		applicationEventsScrapeErrorsTotalMetric:         applicationEventsScrapeErrorsTotalMetric,
		lastApplicationEventsScrapeErrorMetric:           lastApplicationEventsScrapeErrorMetric,
		lastApplicationEventsScrapeTimestampMetric:       lastApplicationEventsScrapeTimestampMetric,
		lastApplicationEventsScrapeDurationSecondsMetric: lastApplicationEventsScrapeDurationSecondsMetric,
	}
}

func (c ApplicationEventsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportApplicationEventsMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.applicationEventsScrapeErrorsTotalMetric.WithLabelValues(c.deploymentName).Inc()
	}
	c.applicationEventsScrapeErrorsTotalMetric.Collect(ch)

	c.applicationEventsScrapesTotalMetric.WithLabelValues(c.deploymentName).Inc()
	c.applicationEventsScrapesTotalMetric.Collect(ch)

	c.lastApplicationEventsScrapeErrorMetric.WithLabelValues(c.deploymentName).Set(errorMetric)
	c.lastApplicationEventsScrapeErrorMetric.Collect(ch)

	c.lastApplicationEventsScrapeTimestampMetric.WithLabelValues(c.deploymentName).Set(float64(time.Now().Unix()))
	c.lastApplicationEventsScrapeTimestampMetric.Collect(ch)

	c.lastApplicationEventsScrapeDurationSecondsMetric.WithLabelValues(c.deploymentName).Set(time.Since(begun).Seconds())
	c.lastApplicationEventsScrapeDurationSecondsMetric.Collect(ch)
}

func (c ApplicationEventsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.applicationEventsTotalMetric.Describe(ch)
	c.applicationEventsScrapesTotalMetric.Describe(ch)
	c.applicationEventsScrapeErrorsTotalMetric.Describe(ch)
	c.lastApplicationEventsScrapeErrorMetric.Describe(ch)
	c.lastApplicationEventsScrapeTimestampMetric.Describe(ch)
	c.lastApplicationEventsScrapeDurationSecondsMetric.Describe(ch)
}

func (c ApplicationEventsCollector) reportApplicationEventsMetrics(ch chan<- prometheus.Metric) error {
	var wg = &sync.WaitGroup{}
	var eventTypes = []string{
		cfclient.AppCrash,
		cfclient.AppCreate,
		cfclient.AppDelete,
		cfclient.AppMapRoute,
		cfclient.AppRestage,
		cfclient.AppSSHAuth,
		cfclient.AppSSHUnauth,
		cfclient.AppStart,
		cfclient.AppStop,
		cfclient.AppUnmapRoute,
		cfclient.AppUpdate,
	}

	c.applicationEventsTotalMetric.Reset()

	doneChannel := make(chan bool, 1)
	errChannel := make(chan error, 1)

	for _, eventType := range eventTypes {
		wg.Add(1)
		go func(eventType string, ch chan<- prometheus.Metric) {
			defer wg.Done()
			if err := c.gatherApplicationEvents(eventType, ch); err != nil {
				errChannel <- err
			}
		}(eventType, ch)
	}

	go func() {
		wg.Wait()
		close(doneChannel)
	}()

	select {
	case <-doneChannel:
	case err := <-errChannel:
		return err
	}

	c.applicationEventsTotalMetric.Collect(ch)

	return nil
}

func (c ApplicationEventsCollector) gatherApplicationEvents(eventType string, ch chan<- prometheus.Metric) error {
	events, err := c.cfClient.ListAppEvents(eventType)
	if err != nil {
		log.Errorf("Error while listing `%s` application events: %v", eventType, err)
		return err
	}

	for _, event := range events {
		if event.ActeeType == "app" {
			c.applicationEventsTotalMetric.WithLabelValues(
				c.deploymentName,
				event.Actee,
				eventType,
			).Inc()
		}
	}

	return nil
}
