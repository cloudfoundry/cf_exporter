package collectors_test

import (
	"flag"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"

	. "github.com/cloudfoundry-community/cf_exporter/collectors"
	. "github.com/cloudfoundry-community/cf_exporter/utils/test_matchers"
)

func init() {
	flag.Set("log.level", "fatal")
}

func applicationEventsHandlerFunc(eventType string, handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v2/events" && r.URL.RawQuery == fmt.Sprintf("q=type:%s", eventType) {
				handler(w, r)
			}
		},
	)
}

var _ = Describe("ApplicationEventsCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		applicationEventsTotalMetric                     *prometheus.CounterVec
		applicationEventsScrapesTotalMetric              prometheus.Counter
		applicationEventsScrapeErrorsTotalMetric         prometheus.Counter
		lastApplicationEventsScrapeErrorMetric           prometheus.Gauge
		lastApplicationEventsScrapeTimestampMetric       prometheus.Gauge
		lastApplicationEventsScrapeDurationSecondsMetric prometheus.Gauge

		namespace      = "test_namespace"
		environment    = "test_environment"
		deployment     = "test_deployment"
		applicationId1 = "fake_application_id_1"
		applicationId2 = "fake_application_id_2"

		applicationEventsCollector *ApplicationEventsCollector
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v2/info"),
				ghttp.RespondWith(http.StatusOK, "{}"),
			),
		)

		cfClient, err = cfclient.NewClient(&cfclient.Config{
			ApiAddress: server.URL(),
			Token:      "fake-token",
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(len(server.ReceivedRequests())).To(Equal(1))

		applicationEventsTotalMetric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "application_events",
				Name:        "total",
				Help:        "Total number of Cloud Foundry Application Events.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"application_id", "event_type"},
		)
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppCrash,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppCreate,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppDelete,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppMapRoute,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppRestage,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppSSHAuth,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppSSHUnauth,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppStart,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppStop,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppUnmapRoute,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId1,
			cfclient.AppUpdate,
		).Inc()
		applicationEventsTotalMetric.WithLabelValues(
			applicationId2,
			cfclient.AppCrash,
		).Inc()

		applicationEventsScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "application_events_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Application Events.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		applicationEventsScrapesTotalMetric.Inc()

		applicationEventsScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "application_events_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape errors of Cloud Foundry Application Events.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastApplicationEventsScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_application_events_scrape_error",
				Help:        "Whether the last scrape of Application Events metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastApplicationEventsScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_application_events_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Application Events metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastApplicationEventsScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_application_events_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Application Events metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		applicationEventsCollector = NewApplicationEventsCollector(namespace, environment, deployment, cfClient)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Describe", func() {
		var (
			descriptions chan *prometheus.Desc
		)

		BeforeEach(func() {
			descriptions = make(chan *prometheus.Desc)
		})

		JustBeforeEach(func() {
			go applicationEventsCollector.Describe(descriptions)
		})

		It("returns an application_events_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppStart,
			).Desc())))
		})

		It("returns an application_events_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(applicationEventsScrapesTotalMetric.Desc())))
		})

		It("returns an application_events_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(applicationEventsScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_application_events_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastApplicationEventsScrapeErrorMetric.Desc())))
		})

		It("returns a last_application_events_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastApplicationEventsScrapeTimestampMetric.Desc())))
		})

		It("returns a last_application_events_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastApplicationEventsScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode                          int
			applicationCrashEventsResponse      cfclient.AppEventResponse
			applicationCreateEventsResponse     cfclient.AppEventResponse
			applicationDeleteEventsResponse     cfclient.AppEventResponse
			applicationMapRouteEventsResponse   cfclient.AppEventResponse
			applicationRestageEventsResponse    cfclient.AppEventResponse
			applicationSSHAuthEventsResponse    cfclient.AppEventResponse
			applicationSSHUnauthEventsResponse  cfclient.AppEventResponse
			applicationStartEventsResponse      cfclient.AppEventResponse
			applicationStopEventsResponse       cfclient.AppEventResponse
			applicationUnmapRouteEventsResponse cfclient.AppEventResponse
			applicationUpdateEventsResponse     cfclient.AppEventResponse

			metrics chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			applicationCrashEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppCrash,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppCrash,
							Actee:     applicationId2,
							ActeeType: "app",
						},
					},
				},
			}
			applicationCrashEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppCrash,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppCrash,
							Actee:     applicationId2,
							ActeeType: "app",
						},
					},
				},
			}
			applicationCreateEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppCreate,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
				},
			}
			applicationDeleteEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppDelete,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
				},
			}
			applicationMapRouteEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppMapRoute,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
				},
			}
			applicationRestageEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppRestage,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
				},
			}
			applicationSSHAuthEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppSSHAuth,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
				},
			}
			applicationSSHUnauthEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppSSHUnauth,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
				},
			}
			applicationStartEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppStart,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
				},
			}
			applicationStopEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppStop,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
				},
			}
			applicationUnmapRouteEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppUnmapRoute,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
				},
			}
			applicationUpdateEventsResponse = cfclient.AppEventResponse{
				Resources: []cfclient.AppEventResource{
					cfclient.AppEventResource{
						Entity: cfclient.AppEventEntity{
							EventType: cfclient.AppUpdate,
							Actee:     applicationId1,
							ActeeType: "app",
						},
					},
				},
			}

			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.RouteToHandler("GET", "/v2/events", ghttp.CombineHandlers(
				applicationEventsHandlerFunc(cfclient.AppCrash, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationCrashEventsResponse)),
				applicationEventsHandlerFunc(cfclient.AppCreate, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationCreateEventsResponse)),
				applicationEventsHandlerFunc(cfclient.AppDelete, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationDeleteEventsResponse)),
				applicationEventsHandlerFunc(cfclient.AppMapRoute, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationMapRouteEventsResponse)),
				applicationEventsHandlerFunc(cfclient.AppRestage, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationRestageEventsResponse)),
				applicationEventsHandlerFunc(cfclient.AppSSHAuth, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationSSHAuthEventsResponse)),
				applicationEventsHandlerFunc(cfclient.AppSSHUnauth, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationSSHUnauthEventsResponse)),
				applicationEventsHandlerFunc(cfclient.AppStart, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationStartEventsResponse)),
				applicationEventsHandlerFunc(cfclient.AppStop, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationStopEventsResponse)),
				applicationEventsHandlerFunc(cfclient.AppUnmapRoute, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationUnmapRouteEventsResponse)),
				applicationEventsHandlerFunc(cfclient.AppUpdate, ghttp.RespondWithJSONEncodedPtr(&statusCode, &applicationUpdateEventsResponse)),
			))

			go applicationEventsCollector.Collect(metrics)
		})

		It("returns an application_events_total metric for app.crash events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppCrash,
			))))
		})

		It("returns an application_events_total metric for app.crash events and application 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppCrash,
			))))
		})

		It("returns an application_events_total metric for audit.app.create events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppCreate,
			))))
		})

		It("does not returns an application_events_total metric for audit.app.create events and application 2", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppCreate,
			))))
		})

		It("returns an application_events_total metric for audit.app.delete-request events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppDelete,
			))))
		})

		It("does not returns an application_events_total metric for audit.app.delete-request events and application 2", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppDelete,
			))))
		})

		It("returns an application_events_total metric for audit.app.map-route events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppMapRoute,
			))))
		})

		It("does not returns an application_events_total metric for audit.app.map-route events and application 2", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppMapRoute,
			))))
		})

		It("returns an application_events_total metric for audit.app.restage events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppRestage,
			))))
		})

		It("does not returns an application_events_total metric for audit.app.restage events and application 2", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppRestage,
			))))
		})
		It("returns an application_events_total metric for audit.app.ssh-authorized events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppSSHAuth,
			))))
		})

		It("does not returns an application_events_total metric for audit.app.ssh-authorized events and application 2", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppSSHAuth,
			))))
		})

		It("returns an application_events_total metric for audit.app.ssh-unauthorized events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppSSHUnauth,
			))))
		})

		It("does not returns an application_events_total metric for audit.app.ssh-unauthorized events and application 2", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppSSHUnauth,
			))))
		})

		It("returns an application_events_total metric for audit.app.start events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppStart,
			))))
		})

		It("does not returns an application_events_total metric for audit.app.start events and application 2", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppStart,
			))))
		})

		It("returns an application_events_total metric for audit.app.stop events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppStop,
			))))
		})

		It("does not returns an application_events_total metric for audit.app.stop events and application 2", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppStop,
			))))
		})

		It("returns an application_events_total metric for audit.app.unmap-route events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppUnmapRoute,
			))))
		})

		It("does not returns an application_events_total metric for audit.app.unmap-route events and application 2", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppUnmapRoute,
			))))
		})

		It("returns an application_events_total metric for audit.app.update events and application 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId1,
				cfclient.AppUpdate,
			))))
		})

		It("does not returns an application_events_total metric for audit.app.update events and application 2", func() {
			Consistently(metrics).ShouldNot(Receive(PrometheusMetric(applicationEventsTotalMetric.WithLabelValues(
				applicationId2,
				cfclient.AppUpdate,
			))))
		})

		It("returns an application_events_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsScrapesTotalMetric)))
		})

		It("returns an application_events_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsScrapeErrorsTotalMetric)))
		})

		It("returns a last_application_events_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastApplicationEventsScrapeErrorMetric)))
		})

		Context("when it fails to list the application events", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				applicationEventsScrapeErrorsTotalMetric.Inc()
				lastApplicationEventsScrapeErrorMetric.Set(1)
			})

			It("returns an application_events_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(applicationEventsScrapeErrorsTotalMetric)))
			})

			It("returns a last_application_events_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastApplicationEventsScrapeErrorMetric)))
			})
		})
	})
})
