package collectors

import (
	"time"

	"github.com/cloudfoundry/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
)

type TasksCollector struct {
	namespace                            string
	environment                          string
	deployment                           string
	taskInfoMetric                       *prometheus.GaugeVec
	tasksCountMetric                     *prometheus.GaugeVec
	tasksMemoryMbSumMetric               *prometheus.GaugeVec
	tasksDiskQuotaMbSumMetric            *prometheus.GaugeVec
	tasksOldestCreatedAtMetric           *prometheus.GaugeVec
	tasksScrapesTotalMetric              prometheus.Counter
	tasksScrapeErrorsTotalMetric         prometheus.Counter
	lastTasksScrapeErrorMetric           prometheus.Gauge
	lastTasksScrapeTimestampMetric       prometheus.Gauge
	lastTasksScrapeDurationSecondsMetric prometheus.Gauge
}

func NewTasksCollector(
	namespace string,
	environment string,
	deployment string,
) *TasksCollector {
	taskInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "task",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Task information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "state"},
	)

	tasksCountMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "task",
			Name:        "count",
			Help:        "Number of Cloud Foundry Tasks.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "state"},
	)

	tasksMemoryMbSumMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "task",
			Name:        "memory_mb_sum",
			Help:        "Sum of Cloud Foundry Tasks Memory (Mb).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "state"},
	)

	tasksDiskQuotaMbSumMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "task",
			Name:        "disk_quota_mb_sum",
			Help:        "Sum of Cloud Foundry Tasks Disk Quota (Mb).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "state"},
	)

	tasksOldestCreatedAtMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "task",
			Name:        "oldest_created_at",
			Help:        "Number of seconds since 1970 of creation time of oldest Cloud Foundry task.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "state"},
	)

	tasksScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "tasks_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Tasks.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	tasksScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "tasks_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape error of Cloud Foundry Tasks.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastTasksScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_tasks_scrape_error",
			Help:        "Whether the last scrape of Tasks metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastTasksScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_tasks_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Tasks metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastTasksScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_tasks_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Tasks metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &TasksCollector{
		namespace:                            namespace,
		environment:                          environment,
		deployment:                           deployment,
		taskInfoMetric:                       taskInfoMetric,
		tasksCountMetric:                     tasksCountMetric,
		tasksMemoryMbSumMetric:               tasksMemoryMbSumMetric,
		tasksDiskQuotaMbSumMetric:            tasksDiskQuotaMbSumMetric,
		tasksOldestCreatedAtMetric:           tasksOldestCreatedAtMetric,
		tasksScrapesTotalMetric:              tasksScrapesTotalMetric,
		tasksScrapeErrorsTotalMetric:         tasksScrapeErrorsTotalMetric,
		lastTasksScrapeErrorMetric:           lastTasksScrapeErrorMetric,
		lastTasksScrapeTimestampMetric:       lastTasksScrapeTimestampMetric,
		lastTasksScrapeDurationSecondsMetric: lastTasksScrapeDurationSecondsMetric,
	}
}

func (c TasksCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.Error != nil {
		errorMetric = float64(1)
		c.tasksScrapeErrorsTotalMetric.Inc()
	} else {
		c.reportTasksMetrics(objs, ch)
	}

	c.tasksScrapeErrorsTotalMetric.Collect(ch)
	c.tasksScrapesTotalMetric.Inc()
	c.tasksScrapesTotalMetric.Collect(ch)

	c.lastTasksScrapeErrorMetric.Set(errorMetric)
	c.lastTasksScrapeErrorMetric.Collect(ch)

	c.lastTasksScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastTasksScrapeTimestampMetric.Collect(ch)
	c.lastTasksScrapeDurationSecondsMetric.Set(objs.Took)

	c.lastTasksScrapeDurationSecondsMetric.Collect(ch)
}

func (c TasksCollector) Describe(ch chan<- *prometheus.Desc) {
	c.taskInfoMetric.Describe(ch)
	c.tasksCountMetric.Describe(ch)
	c.tasksMemoryMbSumMetric.Describe(ch)
	c.tasksDiskQuotaMbSumMetric.Describe(ch)
	c.tasksOldestCreatedAtMetric.Describe(ch)
	c.tasksScrapesTotalMetric.Describe(ch)
	c.tasksScrapeErrorsTotalMetric.Describe(ch)
	c.lastTasksScrapeErrorMetric.Describe(ch)
	c.lastTasksScrapeTimestampMetric.Describe(ch)
	c.lastTasksScrapeDurationSecondsMetric.Describe(ch)
}

func (c TasksCollector) reportTasksMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	c.taskInfoMetric.Reset()
	c.tasksCountMetric.Reset()
	c.tasksMemoryMbSumMetric.Reset()
	c.tasksDiskQuotaMbSumMetric.Reset()
	c.tasksOldestCreatedAtMetric.Reset()

	type keyType struct {
		applicationID string
		state         string
	}
	groupedTasks := map[keyType][]models.Task{}
	for _, task := range objs.Tasks {
		applicationID := "unavailable"
		if app, ok := task.Relationships["app"]; ok && app.GUID != "" {
			applicationID = app.GUID
		}
		key := keyType{applicationID, string(task.State)}

		existingValue, ok := groupedTasks[key]
		if !ok {
			existingValue = []models.Task{}
		}
		groupedTasks[key] = append(existingValue, task)
	}

	for key, tasks := range groupedTasks {
		c.taskInfoMetric.WithLabelValues(
			key.applicationID,
			key.state,
		).Set(float64(1))

		c.tasksCountMetric.WithLabelValues(
			key.applicationID,
			key.state,
		).Set(float64(len(tasks)))

		memorySum := int64(0)
		for _, task := range tasks {
			memorySum += task.MemoryInMb
		}
		c.tasksMemoryMbSumMetric.WithLabelValues(
			key.applicationID,
			key.state,
		).Set(float64(memorySum))

		diskSum := int64(0)
		for _, task := range tasks {
			diskSum += task.DiskInMb
		}
		c.tasksDiskQuotaMbSumMetric.WithLabelValues(
			key.applicationID,
			key.state,
		).Set(float64(diskSum))

		createdAtOldest := time.Now()
		for _, task := range tasks {
			if task.CreatedAt.Before(createdAtOldest) {
				createdAtOldest = task.CreatedAt
			}
		}
		c.tasksOldestCreatedAtMetric.WithLabelValues(
			key.applicationID,
			key.state,
		).Set(float64(createdAtOldest.Unix()))
	}

	c.taskInfoMetric.Collect(ch)
	c.tasksCountMetric.Collect(ch)
	c.tasksMemoryMbSumMetric.Collect(ch)
	c.tasksDiskQuotaMbSumMetric.Collect(ch)
	c.tasksOldestCreatedAtMetric.Collect(ch)
}
