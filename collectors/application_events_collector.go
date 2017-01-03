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
	cfClient                                         *cfclient.Client
	applicationEventsTotalMetric                     *prometheus.CounterVec
	applicationEventsScrapesTotalMetric              prometheus.Counter
	applicationEventsScrapeErrorsTotalMetric         prometheus.Counter
	lastApplicationEventsScrapeErrorMetric           prometheus.Gauge
	lastApplicationEventsScrapeTimestampMetric       prometheus.Gauge
	lastApplicationEventsScrapeDurationSecondsMetric prometheus.Gauge
}

func NewApplicationEventsCollector(namespace string, cfClient *cfclient.Client) *ApplicationEventsCollector {
	applicationEventsTotalMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "application_events",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Application Events.",
		},
		[]string{"application_id", "event_type"},
	)

	applicationEventsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "application_events_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Cloud Foundry Application Events.",
		},
	)

	applicationEventsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "application_events_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrape errors of Cloud Foundry Application Events.",
		},
	)

	lastApplicationEventsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_application_events_scrape_error",
			Help:      "Whether the last scrape of Application Events metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastApplicationEventsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_application_events_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Application Events metrics from Cloud Foundry.",
		},
	)

	lastApplicationEventsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_application_events_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Application Events metrics from Cloud Foundry.",
		},
	)

	return &ApplicationEventsCollector{
		namespace:                                        namespace,
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
		c.applicationEventsScrapeErrorsTotalMetric.Inc()
	}

	c.applicationEventsScrapesTotalMetric.Inc()
	c.applicationEventsScrapesTotalMetric.Collect(ch)

	c.applicationEventsScrapeErrorsTotalMetric.Collect(ch)

	c.lastApplicationEventsScrapeErrorMetric.Set(errorMetric)
	c.lastApplicationEventsScrapeErrorMetric.Collect(ch)

	c.lastApplicationEventsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastApplicationEventsScrapeTimestampMetric.Collect(ch)

	c.lastApplicationEventsScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
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
		cfclient.AppStart,
		cfclient.AppStop,
		cfclient.AppUpdate,
		cfclient.AppCreate,
		cfclient.AppDelete,
		cfclient.AppSSHAuth,
		cfclient.AppSSHUnauth,
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
				event.Actee,
				eventType,
			).Inc()
		}
	}

	return nil
}
