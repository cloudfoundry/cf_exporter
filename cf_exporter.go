package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"

	"github.com/cloudfoundry-community/cf_exporter/collectors"
	"github.com/cloudfoundry-community/cf_exporter/filters"
)

var (
	cfAPIUrl = flag.String(
		"cf.api_url", "",
		"Cloud Foundry API URL ($CF_EXPORTER_CF_API_URL).",
	)

	cfUsername = flag.String(
		"cf.username", "",
		"Cloud Foundry Username ($CF_EXPORTER_CF_USERNAME).",
	)

	cfPassword = flag.String(
		"cf.password", "",
		"Cloud Foundry Password ($CF_EXPORTER_CF_PASSWORD).",
	)

	cfClientID = flag.String(
		"cf.client-id", "",
		"Cloud Foundry Client ID ($CF_EXPORTER_CF_CLIENT_ID).",
	)

	cfClientSecret = flag.String(
		"cf.client-secret", "",
		"Cloud Foundry Client Secret ($CF_EXPORTER_CF_CLIENT_SECRET).",
	)

	cfDeploymentName = flag.String(
		"cf.deployment-name", "",
		"Cloud Foundry Deployment Name to be reported as a metric label ($CF_EXPORTER_CF_DEPLOYMENT_NAME).",
	)

	filterCollectors = flag.String(
		"filter.collectors", "",
		"Comma separated collectors to filter (Applications,ApplicationEvents,Organizations,SecurityGroups,Services,Spaces) ($CF_EXPORTER_FILTER_COLLECTORS).",
	)

	metricsNamespace = flag.String(
		"metrics.namespace", "cf",
		"Metrics Namespace ($CF_EXPORTER_METRICS_NAMESPACE).",
	)

	showVersion = flag.Bool(
		"version", false,
		"Print version information.",
	)

	skipSSLValidation = flag.Bool(
		"skip-ssl-verify", false,
		"Disable SSL Verify ($CF_EXPORTER_SKIP_SSL_VERIFY).",
	)

	listenAddress = flag.String(
		"web.listen-address", ":9193",
		"Address to listen on for web interface and telemetry ($CF_EXPORTER_WEB_LISTEN_ADDRESS).",
	)

	metricsPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose Prometheus metrics ($CF_EXPORTER_WEB_TELEMETRY_PATH).",
	)

	authUsername = flag.String(
		"web.auth.username", "",
		"Username for web interface basic auth ($CF_EXPORTER_WEB_AUTH_USERNAME).",
	)

	authPassword = flag.String(
		"web.auth.password", "",
		"Password for web interface basic auth ($CF_EXPORTER_WEB_AUTH_PASSWORD).",
	)

	tlsCertFile = flag.String(
		"web.tls.cert_file", "",
		"Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate ($CF_EXPORTER_WEB_TLS_CERTFILE).",
	)

	tlsKeyFile = flag.String(
		"web.tls.key_file", "",
		"Path to a file that contains the TLS private key (PEM format) ($CF_EXPORTER_WEB_TLS_KEYFILE).",
	)
)

func init() {
	prometheus.MustRegister(version.NewCollector(*metricsNamespace))
}

func overrideFlagsWithEnvVars() {
	overrideWithEnvVar("CF_EXPORTER_CF_API_URL", cfAPIUrl)
	overrideWithEnvVar("CF_EXPORTER_CF_USERNAME", cfUsername)
	overrideWithEnvVar("CF_EXPORTER_CF_PASSWORD", cfPassword)
	overrideWithEnvVar("CF_EXPORTER_CF_CLIENT_ID", cfClientID)
	overrideWithEnvVar("CF_EXPORTER_CF_CLIENT_SECRET", cfClientSecret)
	overrideWithEnvVar("CF_EXPORTER_CF_DEPLOYMENT_NAME", cfDeploymentName)
	overrideWithEnvVar("CF_EXPORTER_FILTER_COLLECTORS", filterCollectors)
	overrideWithEnvVar("CF_EXPORTER_METRICS_NAMESPACE", metricsNamespace)
	overrideWithEnvBool("CF_EXPORTER_SKIP_SSL_VERIFY", skipSSLValidation)
	overrideWithEnvVar("CF_EXPORTER_WEB_LISTEN_ADDRESS", listenAddress)
	overrideWithEnvVar("CF_EXPORTER_WEB_TELEMETRY_PATH", metricsPath)
	overrideWithEnvVar("CF_EXPORTER_WEB_AUTH_USERNAME", authUsername)
	overrideWithEnvVar("CF_EXPORTER_WEB_AUTH_PASSWORD", authPassword)
	overrideWithEnvVar("CF_EXPORTER_WEB_TLS_CERTFILE", tlsCertFile)
	overrideWithEnvVar("CF_EXPORTER_WEB_TLS_KEYFILE", tlsKeyFile)
}

func overrideWithEnvVar(name string, value *string) {
	envValue := os.Getenv(name)
	if envValue != "" {
		*value = envValue
	}
}

func overrideWithEnvBool(name string, value *bool) {
	envValue := os.Getenv(name)
	if envValue != "" {
		var err error
		*value, err = strconv.ParseBool(envValue)
		if err != nil {
			log.Fatalf("Invalid `%s`: %s", name, err)
		}
	}
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
	flag.Parse()
	overrideFlagsWithEnvVars()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("cf_exporter"))
		os.Exit(0)
	}

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
		applicationsCollector := collectors.NewApplicationsCollector(*metricsNamespace, *cfDeploymentName, cfClient)
		prometheus.MustRegister(applicationsCollector)
	}

	if collectorsFilter.Enabled(filters.ApplicationEventsCollector) {
		applicationEventsCollector := collectors.NewApplicationEventsCollector(*metricsNamespace, *cfDeploymentName, cfClient)
		prometheus.MustRegister(applicationEventsCollector)
	}

	if collectorsFilter.Enabled(filters.OrganizationsCollector) {
		organizationsCollector := collectors.NewOrganizationsCollector(*metricsNamespace, *cfDeploymentName, cfClient)
		prometheus.MustRegister(organizationsCollector)
	}

	if collectorsFilter.Enabled(filters.SecurityGroupsCollector) {
		securityGroupsCollector := collectors.NewSecurityGroupsCollector(*metricsNamespace, *cfDeploymentName, cfClient)
		prometheus.MustRegister(securityGroupsCollector)
	}

	if collectorsFilter.Enabled(filters.ServicesCollector) {
		servicesCollector := collectors.NewServicesCollector(*metricsNamespace, *cfDeploymentName, cfClient)
		prometheus.MustRegister(servicesCollector)
	}

	if collectorsFilter.Enabled(filters.SpacesCollector) {
		spacesCollector := collectors.NewSpacesCollector(*metricsNamespace, *cfDeploymentName, cfClient)
		prometheus.MustRegister(spacesCollector)
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
