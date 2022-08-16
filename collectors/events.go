package collectors

import (
	"time"

	"github.com/bosh-prometheus/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type EventsCollector struct {
	namespace                             string
	environment                           string
	deployment                            string
	eventsInfoMetric                      *prometheus.GaugeVec
	eventsScrapesTotalMetric              prometheus.Counter
	eventsScrapeErrorsTotalMetric         prometheus.Counter
	lastEventsScrapeErrorMetric           prometheus.Gauge
	lastEventsScrapeTimestampMetric       prometheus.Gauge
	lastEventsScrapeDurationSecondsMetric prometheus.Gauge
	lastCheckFilter time.Time
	timeLocation    *time.Location
}

func NewEventsCollector(
	namespace string,
	environment string,
	deployment string,
) *EventsCollector {
	eventsInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "events",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Events information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"type", "actor", "actor_type", "actor_name", "actor_username", "actee", "actee_type", "actee_name", "space_id", "organization_id"},
	)

	eventsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "events_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Events.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	eventsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "events_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape errors of Cloud Foundry Events.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastEventsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_events_scrape_error",
			Help:        "Whether the last scrape of Event metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastEventsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_events_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Event metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastEventsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_events_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Events metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	timeLocation, _ := time.LoadLocation("UTC")
	now := time.Now().In(timeLocation)

	return &EventsCollector{
		namespace:                             namespace,
		environment:                           environment,
		deployment:                            deployment,
		eventsInfoMetric:                      eventsInfoMetric,
		eventsScrapesTotalMetric:              eventsScrapesTotalMetric,
		eventsScrapeErrorsTotalMetric:         eventsScrapeErrorsTotalMetric,
		lastEventsScrapeErrorMetric:           lastEventsScrapeErrorMetric,
		lastEventsScrapeTimestampMetric:       lastEventsScrapeTimestampMetric,
		lastEventsScrapeDurationSecondsMetric: lastEventsScrapeDurationSecondsMetric,
		lastCheckFilter:                       now,
		timeLocation:                          timeLocation,
	}
}

func (c *EventsCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	err := objs.Error
	if objs.Error != nil {
		errorMetric = float64(1)
		c.eventsScrapeErrorsTotalMetric.Inc()
	} else {
		err = c.reportEventsMetrics(objs, ch)
		if err != nil {
			log.Error(err)
			errorMetric = float64(1)
			c.eventsScrapeErrorsTotalMetric.Inc()
		}
	}

	c.eventsScrapeErrorsTotalMetric.Collect(ch)
	c.eventsScrapesTotalMetric.Inc()
	c.eventsScrapesTotalMetric.Collect(ch)
	c.lastEventsScrapeErrorMetric.Set(errorMetric)
	c.lastEventsScrapeErrorMetric.Collect(ch)
	c.lastEventsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastEventsScrapeTimestampMetric.Collect(ch)
	c.lastEventsScrapeDurationSecondsMetric.Set(objs.Took)
	c.lastEventsScrapeDurationSecondsMetric.Collect(ch)
}

func (c *EventsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.eventsInfoMetric.Describe(ch)
	c.eventsScrapesTotalMetric.Describe(ch)
	c.eventsScrapeErrorsTotalMetric.Describe(ch)
	c.lastEventsScrapeErrorMetric.Describe(ch)
	c.lastEventsScrapeTimestampMetric.Describe(ch)
	c.lastEventsScrapeDurationSecondsMetric.Describe(ch)
}

// reportEventsMetrics
// 1. find user's username in user map if available
func (c *EventsCollector) reportEventsMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) error {
	c.eventsInfoMetric.Reset()

	for _, event := range objs.Events {
		if event.CreatedAt.Before(c.lastCheckFilter) {
			continue
		}

		// 1.
		actorUsername := ""
		if user, ok := objs.Users[event.Actor.GUID]; ok {
			actorUsername = user.Username
		}

		c.eventsInfoMetric.WithLabelValues(
			event.Type,
			event.Actor.GUID,
			event.Actor.Type,
			event.Actor.Name,
			actorUsername,
			event.Target.GUID,
			event.Target.Type,
			event.Target.Name,
			event.Space.GUID,
			event.Org.GUID,
		).Set(float64(1))
	}

	timeLocation, _ := time.LoadLocation("UTC")
	c.lastCheckFilter = time.Now().In(timeLocation)
	c.eventsInfoMetric.Collect(ch)
	return nil
}
