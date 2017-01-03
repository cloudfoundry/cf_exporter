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
		lastApplicationEventsScrapeErrorMetric:           lastApplicationEventsScrapeErrorMetric,
		lastApplicationEventsScrapeTimestampMetric:       lastApplicationEventsScrapeTimestampMetric,
		lastApplicationEventsScrapeDurationSecondsMetric: lastApplicationEventsScrapeDurationSecondsMetric,
	}
}

func (c ApplicationEventsCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()
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
		c.reportErrorMetric(false, ch)
	case <-errChannel:
		c.reportErrorMetric(true, ch)
	}

	c.applicationEventsTotalMetric.Collect(ch)

	c.lastApplicationEventsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastApplicationEventsScrapeTimestampMetric.Collect(ch)

	c.lastApplicationEventsScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastApplicationEventsScrapeDurationSecondsMetric.Collect(ch)
}

func (c ApplicationEventsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.applicationEventsTotalMetric.Describe(ch)
	c.lastApplicationEventsScrapeErrorMetric.Describe(ch)
	c.lastApplicationEventsScrapeTimestampMetric.Describe(ch)
	c.lastApplicationEventsScrapeDurationSecondsMetric.Describe(ch)
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

func (c ApplicationEventsCollector) reportErrorMetric(errorHappend bool, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if errorHappend {
		errorMetric = float64(1)
	}

	c.lastApplicationEventsScrapeErrorMetric.Set(errorMetric)
	c.lastApplicationEventsScrapeErrorMetric.Collect(ch)
}
