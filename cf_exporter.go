package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/bosh-prometheus/cf_exporter/collectors"
	"github.com/bosh-prometheus/cf_exporter/filters"
)

var (
	cfAPIUrl = kingpin.Flag(
		"cf.api_url", "Cloud Foundry API URL ($CF_EXPORTER_CF_API_URL)",
	).Envar("CF_EXPORTER_CF_API_URL").Required().String()

	cfUsername = kingpin.Flag(
		"cf.username", "Cloud Foundry Username ($CF_EXPORTER_CF_USERNAME)",
	).Envar("CF_EXPORTER_CF_USERNAME").String()

	cfPassword = kingpin.Flag(
		"cf.password", "Cloud Foundry Password ($CF_EXPORTER_CF_PASSWORD)",
	).Envar("CF_EXPORTER_CF_PASSWORD").String()

	cfClientID = kingpin.Flag(
		"cf.client-id", "Cloud Foundry Client ID ($CF_EXPORTER_CF_CLIENT_ID)",
	).Envar("CF_EXPORTER_CF_CLIENT_ID").String()

	cfClientSecret = kingpin.Flag(
		"cf.client-secret", "Cloud Foundry Client Secret ($CF_EXPORTER_CF_CLIENT_SECRET)",
	).Envar("CF_EXPORTER_CF_CLIENT_SECRET").String()

	cfDeploymentName = kingpin.Flag(
		"cf.deployment-name", "Cloud Foundry Deployment Name to be reported as a metric label ($CF_EXPORTER_CF_DEPLOYMENT_NAME)",
	).Envar("CF_EXPORTER_CF_DEPLOYMENT_NAME").Required().String()

	filterCollectors = kingpin.Flag(
		"filter.collectors", "Comma separated collectors to filter (Applications,IsolationSegments,Organizations,Routes,SecurityGroups,ServiceBindings,ServiceInstances,ServicePlans,Services,Spaces,Stacks) ($CF_EXPORTER_FILTER_COLLECTORS)",
	).Envar("CF_EXPORTER_FILTER_COLLECTORS").Default("").String()

	metricsNamespace = kingpin.Flag(
		"metrics.namespace", "Metrics Namespace ($CF_EXPORTER_METRICS_NAMESPACE)",
	).Envar("CF_EXPORTER_METRICS_NAMESPACE").Default("cf").String()

	metricsEnvironment = kingpin.Flag(
		"metrics.environment", "Environment label to be attached to metrics ($CF_EXPORTER_METRICS_ENVIRONMENT)",
	).Envar("CF_EXPORTER_METRICS_ENVIRONMENT").Required().String()

	skipSSLValidation = kingpin.Flag(
		"skip-ssl-verify", "Disable SSL Verify ($CF_EXPORTER_SKIP_SSL_VERIFY)",
	).Envar("CF_EXPORTER_SKIP_SSL_VERIFY").Default("false").Bool()

	listenAddress = kingpin.Flag(
		"web.listen-address", "Address to listen on for web interface and telemetry ($CF_EXPORTER_WEB_LISTEN_ADDRESS)",
	).Envar("CF_EXPORTER_WEB_LISTEN_ADDRESS").Default(":9193").String()

	metricsPath = kingpin.Flag(
		"web.telemetry-path", "Path under which to expose Prometheus metrics ($CF_EXPORTER_WEB_TELEMETRY_PATH)",
	).Envar("CF_EXPORTER_WEB_TELEMETRY_PATH").Default("/metrics").String()

	authUsername = kingpin.Flag(
		"web.auth.username", "Username for web interface basic auth ($CF_EXPORTER_WEB_AUTH_USERNAME)",
	).Envar("CF_EXPORTER_WEB_AUTH_USERNAME").String()

	authPassword = kingpin.Flag(
		"web.auth.password", "Password for web interface basic auth ($CF_EXPORTER_WEB_AUTH_PASSWORD)",
	).Envar("CF_EXPORTER_WEB_AUTH_PASSWORD").String()

	tlsCertFile = kingpin.Flag(
		"web.tls.cert_file", "Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate ($CF_EXPORTER_WEB_TLS_CERTFILE)",
	).Envar("CF_EXPORTER_WEB_TLS_KEYFILE").ExistingFile()

	tlsKeyFile = kingpin.Flag(
		"web.tls.key_file", "Path to a file that contains the TLS private key (PEM format) ($CF_EXPORTER_WEB_TLS_KEYFILE)",
	).Envar("CF_EXPORTER_WEB_TLS_KEYFILE").ExistingFile()
)

func init() {
	prometheus.MustRegister(version.NewCollector(*metricsNamespace))
}

type basicAuthHandler struct {
	handler  http.HandlerFunc
	username string
	password string
}

func (h *basicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || username != h.username || password != h.password {
		log.Errorf("Invalid HTTP auth from `%s`", r.RemoteAddr)
		w.Header().Set("WWW-Authenticate", "Basic realm=\"metrics\"")
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	h.handler(w, r)
	return
}

func prometheusHandler() http.Handler {
	handler := prometheus.Handler()

	if *authUsername != "" && *authPassword != "" {
		handler = &basicAuthHandler{
			handler:  prometheus.Handler().ServeHTTP,
			username: *authUsername,
			password: *authPassword,
		}
	}

	return handler
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("cf_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting cf_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	cfConfig := &cfclient.Config{
		ApiAddress:        *cfAPIUrl,
		Username:          *cfUsername,
		Password:          *cfPassword,
		ClientID:          *cfClientID,
		ClientSecret:      *cfClientSecret,
		SkipSslValidation: *skipSSLValidation,
		UserAgent:         fmt.Sprintf("cf_exporter/%s", version.Version),
	}
	cfClient, err := cfclient.NewClient(cfConfig)
	if err != nil {
		log.Errorf("Error creating Cloud Foundry client: %s", err.Error())
		os.Exit(1)
	}

	var collectorsFilters []string
	if *filterCollectors != "" {
		collectorsFilters = strings.Split(*filterCollectors, ",")
	}
	collectorsFilter, err := filters.NewCollectorsFilter(collectorsFilters)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if collectorsFilter.Enabled(filters.ApplicationsCollector) {
		applicationsCollector := collectors.NewApplicationsCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(applicationsCollector)
	}

	if collectorsFilter.Enabled(filters.IsolationSegmentsCollector) {
		isolationSegmentsCollector := collectors.NewIsolationSegmentsCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(isolationSegmentsCollector)
	}

	if collectorsFilter.Enabled(filters.OrganizationsCollector) {
		organizationsCollector := collectors.NewOrganizationsCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(organizationsCollector)
	}

	if collectorsFilter.Enabled(filters.RoutesCollector) {
		routesCollector := collectors.NewRoutesCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(routesCollector)
	}

	if collectorsFilter.Enabled(filters.SecurityGroupsCollector) {
		securityGroupsCollector := collectors.NewSecurityGroupsCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(securityGroupsCollector)
	}

	if collectorsFilter.Enabled(filters.ServiceBindingsCollector) {
		serviceBindingsCollector := collectors.NewServiceBindingsCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(serviceBindingsCollector)
	}

	if collectorsFilter.Enabled(filters.ServiceInstancesCollector) {
		serviceInstancesCollector := collectors.NewServiceInstancesCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(serviceInstancesCollector)
	}

	if collectorsFilter.Enabled(filters.ServicePlansCollector) {
		servicePlansCollector := collectors.NewServicePlansCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(servicePlansCollector)
	}

	if collectorsFilter.Enabled(filters.ServicesCollector) {
		servicesCollector := collectors.NewServicesCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(servicesCollector)
	}

	if collectorsFilter.Enabled(filters.SpacesCollector) {
		spacesCollector := collectors.NewSpacesCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(spacesCollector)
	}

	if collectorsFilter.Enabled(filters.StacksCollector) {
		stacksCollector := collectors.NewStacksCollector(*metricsNamespace, *metricsEnvironment, *cfDeploymentName, cfClient)
		prometheus.MustRegister(stacksCollector)
	}

	handler := prometheusHandler()
	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Cloud Foundry Exporter</title></head>
             <body>
             <h1>Cloud Foundry Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	if *tlsCertFile != "" && *tlsKeyFile != "" {
		log.Infoln("Listening TLS on", *listenAddress)
		log.Fatal(http.ListenAndServeTLS(*listenAddress, *tlsCertFile, *tlsKeyFile, nil))
	} else {
		log.Infoln("Listening on", *listenAddress)
		log.Fatal(http.ListenAndServe(*listenAddress, nil))
	}
}
