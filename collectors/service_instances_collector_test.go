package collectors_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	. "github.com/bosh-prometheus/cf_exporter/collectors"
	. "github.com/bosh-prometheus/cf_exporter/utils/test_matchers"
)

func init() {
	log.Base().SetLevel("fatal")
}

var _ = Describe("ServiceIntancesCollectors", func() {
	var (
		err      error
		server   *ghttp.Server
		cfClient *cfclient.Client

		serviceInstanceInfoMetric                       *prometheus.GaugeVec
		serviceInstancesScrapesTotalMetric              prometheus.Counter
		serviceInstancesScrapeErrorsTotalMetric         prometheus.Counter
		lastServiceInstancesScrapeErrorMetric           prometheus.Gauge
		lastServiceInstancesScrapeTimestampMetric       prometheus.Gauge
		lastServiceInstancesScrapeDurationSecondsMetric prometheus.Gauge

		namespace                          = "test_namespace"
		environment                        = "test_environment"
		deployment                         = "test_deployment"
		serviceInstanceId1                 = "fake_service_instance_id_1"
		serviceInstanceName1               = "fake_service_instance_name_1"
		serviceInstanceServicePlanId1      = "fake_service_instance_service_plan_id_1"
		serviceInstanceSpaceId1            = "fake_service_instance_space_id_1"
		serviceInstanceType1               = "fake_service_instance_type_1"
		serviceInstanceLastOperationType1  = "fake_service_instance_last_operation_type_1"
		serviceInstanceLastOperationState1 = "fake_service_instance_last_operation_state_1"
		serviceInstanceId2                 = "fake_service_instance_id_2"
		serviceInstanceName2               = "fake_service_instance_name_2"
		serviceInstanceServicePlanId2      = "fake_service_instance_service_plan_id_2"
		serviceInstanceSpaceId2            = "fake_service_instance_space_id_2"
		serviceInstanceType2               = "fake_service_instance_type_2"
		serviceInstanceLastOperationType2  = "fake_service_instance_last_operation_type_2"
		serviceInstanceLastOperationState2 = "fake_service_instance_last_operation_state_2"

		serviceInstancesCollector *ServiceInstancesCollector
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

		serviceInstanceInfoMetric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "service_instance",
				Name:        "info",
				Help:        "Labeled Cloud Foundry Service Instance information with a constant '1' value.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
			[]string{"service_instance_id", "service_instance_name", "service_plan_id", "space_id", "type", "last_operation_type", "last_operation_state"},
		)
		serviceInstanceInfoMetric.WithLabelValues(
			serviceInstanceId1,
			serviceInstanceName1,
			serviceInstanceServicePlanId1,
			serviceInstanceSpaceId1,
			serviceInstanceType1,
			serviceInstanceLastOperationType1,
			serviceInstanceLastOperationState1,
		).Set(1)
		serviceInstanceInfoMetric.WithLabelValues(
			serviceInstanceId2,
			serviceInstanceName2,
			serviceInstanceServicePlanId2,
			serviceInstanceSpaceId2,
			serviceInstanceType2,
			serviceInstanceLastOperationType2,
			serviceInstanceLastOperationState2,
		).Set(1)

		serviceInstancesScrapesTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "service_instances_scrapes",
				Name:        "total",
				Help:        "Total number of scrapes for Cloud Foundry Service Instances.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
		serviceInstancesScrapesTotalMetric.Inc()

		serviceInstancesScrapeErrorsTotalMetric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "service_instances_scrape_errors",
				Name:        "total",
				Help:        "Total number of scrape error of Cloud Foundry Service Instances.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServiceInstancesScrapeErrorMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_service_instances_scrape_error",
				Help:        "Whether the last scrape of Service Instances metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServiceInstancesScrapeTimestampMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_service_instances_scrape_timestamp",
				Help:        "Number of seconds since 1970 since last scrape of Service Instances metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)

		lastServiceInstancesScrapeDurationSecondsMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "",
				Name:        "last_service_instances_scrape_duration_seconds",
				Help:        "Duration of the last scrape of Service Instances metrics from Cloud Foundry.",
				ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
			},
		)
	})

	JustBeforeEach(func() {
		serviceInstancesCollector = NewServiceInstancesCollector(namespace, environment, deployment, cfClient)
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
			go serviceInstancesCollector.Describe(descriptions)
		})

		It("returns a service_instance_info metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(serviceInstanceInfoMetric.WithLabelValues(
				serviceInstanceId1,
				serviceInstanceName1,
				serviceInstanceServicePlanId1,
				serviceInstanceSpaceId1,
				serviceInstanceType1,
				serviceInstanceLastOperationType1,
				serviceInstanceLastOperationState1,
			).Desc())))
		})

		It("returns a service_instances_scrapes_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(serviceInstancesScrapesTotalMetric.Desc())))
		})

		It("returns a service_instances_scrape_errors_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(serviceInstancesScrapeErrorsTotalMetric.Desc())))
		})

		It("returns a last_service_instances_scrape_error metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServiceInstancesScrapeErrorMetric.Desc())))
		})

		It("returns a last_service_instances_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServiceInstancesScrapeTimestampMetric.Desc())))
		})

		It("returns a last_service_instances_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastServiceInstancesScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			statusCode               int
			serviceInstancesResponse cfclient.ServiceInstancesResponse
			metrics                  chan prometheus.Metric
		)

		BeforeEach(func() {
			statusCode = http.StatusOK
			serviceInstancesResponse = cfclient.ServiceInstancesResponse{
				Resources: []cfclient.ServiceInstanceResource{
					cfclient.ServiceInstanceResource{
						Meta: cfclient.Meta{
							Guid: serviceInstanceId1,
						},
						Entity: cfclient.ServiceInstance{
							Name:            serviceInstanceName1,
							ServicePlanGuid: serviceInstanceServicePlanId1,
							SpaceGuid:       serviceInstanceSpaceId1,
							Type:            serviceInstanceType1,
							LastOperation: cfclient.LastOperation{
								Type:  serviceInstanceLastOperationType1,
								State: serviceInstanceLastOperationState1,
							},
						},
					},
					cfclient.ServiceInstanceResource{
						Meta: cfclient.Meta{
							Guid: serviceInstanceId2,
						},
						Entity: cfclient.ServiceInstance{
							Name:            serviceInstanceName2,
							ServicePlanGuid: serviceInstanceServicePlanId2,
							SpaceGuid:       serviceInstanceSpaceId2,
							Type:            serviceInstanceType2,
							LastOperation: cfclient.LastOperation{
								Type:  serviceInstanceLastOperationType2,
								State: serviceInstanceLastOperationState2,
							},
						},
					},
				},
			}
			metrics = make(chan prometheus.Metric)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/service_instances"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &serviceInstancesResponse),
				),
			)

			go serviceInstancesCollector.Collect(metrics)
		})

		It("returns a service_instance_info metric for service instance 1", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(serviceInstanceInfoMetric.WithLabelValues(
				serviceInstanceId1,
				serviceInstanceName1,
				serviceInstanceServicePlanId1,
				serviceInstanceSpaceId1,
				serviceInstanceType1,
				serviceInstanceLastOperationType1,
				serviceInstanceLastOperationState1,
			))))
		})

		It("returns a service_instance_info metric for service instance 2", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(serviceInstanceInfoMetric.WithLabelValues(
				serviceInstanceId2,
				serviceInstanceName2,
				serviceInstanceServicePlanId2,
				serviceInstanceSpaceId2,
				serviceInstanceType2,
				serviceInstanceLastOperationType2,
				serviceInstanceLastOperationState2,
			))))
		})

		It("returns a service_instances_scrapes_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(serviceInstancesScrapesTotalMetric)))
		})

		It("returns a service_instances_scrape_errors_total metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(serviceInstancesScrapeErrorsTotalMetric)))
		})

		It("returns a last_service_instances_scrape_error metric", func() {
			Eventually(metrics).Should(Receive(PrometheusMetric(lastServiceInstancesScrapeErrorMetric)))
		})

		Context("when it fails to list the service instances", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError

				serviceInstancesScrapeErrorsTotalMetric.Inc()
				lastServiceInstancesScrapeErrorMetric.Set(1)
			})

			It("returns a service_instances_scrape_errors_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(serviceInstancesScrapeErrorsTotalMetric)))
			})

			It("returns a last_service_instances_scrape_error metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(lastServiceInstancesScrapeErrorMetric)))
			})
		})
	})
})
