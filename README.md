# Cloud Foundry Prometheus Exporter [![Build Status](https://travis-ci.org/cloudfoundry-community/cf_exporter.png)](https://travis-ci.org/cloudfoundry-community/cf_exporter)

A [Prometheus][prometheus] exporter for [Cloud Foundry][cloudfoundry] administrative information. Please refer to the [FAQ][faq] for general questions about this exporter.

## Architecture overview

![](https://cdn.rawgit.com/cloudfoundry-community/cf_exporter/master/architecture/architecture.svg)

## Installation

### Binaries

Download the already existing [binaries][binaries] for your platform:

```bash
$ ./cf_exporter <flags>
```

### From source

Using the standard `go install` (you must have [Go][golang] already installed in your local machine):

```bash
$ go install github.com/cloudfoundry-community/cf_exporter
$ cf_exporter <flags>
```

### Cloud Foundry

The exporter can be deployed to an already existing [Cloud Foundry][cloudfoundry] environment:

```bash
$ git clone https://github.com/cloudfoundry-community/cf_exporter.git
$ cd cf_exporter
```

Modify the included [application manifest file][manifest] to include your Cloud Foundry properties. Then you can push the exporter to your Cloud Foundry environment:

```bash
$ cf push
```

### BOSH

This exporter can be deployed using the [Prometheus BOSH Release][prometheus-boshrelease].

## Usage

### Flags

| Flag / Environment Variable | Required | Default | Description |
| --------------------------- | -------- | ------- | ----------- |
| `cf.api_url`<br />`CF_EXPORTER_CF_API_URL` | Yes | | Cloud Foundry API URL |
| `cf.username`<br />`CF_EXPORTER_CF_USERNAME` | *[1]* | | Cloud Foundry Username (must be an `administrator` user) |
| `cf.password`<br />`CF_EXPORTER_CF_PASSWORD` | *[1]* | | Cloud Foundry Password |
| `cf.client-id`<br />`CF_EXPORTER_CF_CLIENT_ID` | *[1]* | | Cloud Foundry Client ID (must have the `cloud_controller.admin` or `cloud_controller.admin_read_only` scope) |
| `cf.client-secret`<br />`CF_EXPORTER_CF_CLIENT_SECRET` | *[1]* | | Cloud Foundry Client Secret |
| `cf.deployment-name`<br />`CF_EXPORTER_CF_DEPLOYMENT_NAME` | No | | Cloud Foundry Deployment Name to be reported as a metric label |
| `filter.collectors`<br />`CF_EXPORTER_FILTER_COLLECTORS` | No | | Comma separated collectors to filter. If not set, all collectors will be enabled (`Applications`, `ApplicationEvents`, `Organizations`, `Routes`, `SecurityGroups`, `ServiceInstances`, `Services`, `Spaces`, `Stacks`) |
| `metrics.namespace`<br />`CF_EXPORTER_METRICS_NAMESPACE` | No | `cf` | Metrics Namespace |
| `metrics.environment`<br />`CF_EXPORTER_METRICS_ENVIRONMENT` | No | | Environment label to be attached to metrics |
| `skip-ssl-verify`<br />`CF_EXPORTER_SKIP_SSL_VERIFY` | No | `false` | Disable SSL Verify |
| `web.listen-address`<br />`CF_EXPORTER_WEB_LISTEN_ADDRESS` | No | `:9193` | Address to listen on for web interface and telemetry |
| `web.telemetry-path`<br />`CF_EXPORTER_WEB_TELEMETRY_PATH` | No | `/metrics` | Path under which to expose Prometheus metrics |
| `web.auth.username`<br />`CF_EXPORTER_WEB_AUTH_USERNAME` | No | | Username for web interface basic auth |
| `web.auth.pasword`<br />`CF_EXPORTER_WEB_AUTH_PASSWORD` | No | | Password for web interface basic auth |
| `web.tls.cert_file`<br />`CF_EXPORTER_WEB_TLS_CERTFILE` | No | | Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate |
| `web.tls.key_file`<br />`CF_EXPORTER_WEB_TLS_KEYFILE` | No | | Path to a file that contains the TLS private key (PEM format) |

*[1]* Either `cf.username` and `cf.password` or `cf.client-id` and `cf.client-secret` must be provided, depending on the desired authentication method.

### Metrics

The exporter returns the following `Applications` metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_application_info | Labeled Cloud Foundry Application information with a constant `1` value | `environment`, `deployment`, `application_id`, `application_name`, `buildpack`, `organization_id`, `organization_name`, `space_id`, `space_name`, `stack_id`, `state` |
| *metrics.namespace*_application_instances | Cloud Foundry Application Instances | `environment`, `deployment`, `application_id`, `application_name`, `organization_id`, `organization_name`, `space_id`, `space_name` |
| *metrics.namespace*_application_memory_mb | Cloud Foundry Application Memory (Mb) | `environment`, `deployment`, `application_id`, `application_name`, `organization_id`, `organization_name`, `space_id`, `space_name` |
| *metrics.namespace*_application_disk_quota_mb | Cloud Foundry Application Disk Quota (Mb) | `environment`, `deployment`, `application_id`, `application_name`, `organization_id`, `organization_name`, `space_id`, `space_name` |
| *metrics.namespace*_applications_scrapes_total | Total number of scrapes for Cloud Foundry Applications | `environment`, `deployment` |
| *metrics.namespace*_applications_scrape_errors_total | Total number of scrape errors of Cloud Foundry Applications | `environment`, `deployment` |
| *metrics.namespace*_last_applications_scrape_error | Whether the last scrape of Applications metrics from Cloud Foundry resulted in an error (`1` for error, `0` for success) | `environment`, `deployment` |
| *metrics.namespace*_last_applications_scrape_timestamp | Number of seconds since 1970 since last scrape of Applications metrics from Cloud Foundry | `environment`, `deployment` |
| *metrics.namespace*_last_applications_scrape_duration_seconds | Duration of the last scrape of Applications metrics from Cloud Foundry | `environment`, `deployment` |

The exporter returns the following `Application Events` metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_application_events_total | Total number of Cloud Foundry Application Events | `environment`, `deployment`, `application_id`, `event_type` |
| *metrics.namespace*_application_events_scrapes_total | Total number of scrapes for Cloud Foundry Application Events | `environment`, `deployment` |
| *metrics.namespace*_application_events_scrape_errors_total | Total number of scrape errors of Cloud Foundry Application Events | `environment`, `deployment` |
| *metrics.namespace*_last_application_events_scrape_error | Whether the last scrape of Application Events metrics from Cloud Foundry resulted in an error (`1` for error, `0` for success) |`environment`,  `deployment` |
| *metrics.namespace*_last_application_events_scrape_timestamp | Number of seconds since 1970 since last scrape of Application Events metrics from Cloud Foundry | `environment`, `deployment` |
| *metrics.namespace*_last_application_events_scrape_duration_seconds | Duration of the last scrape of Application Events metrics from Cloud Foundry | `environment`, `deployment` |

The exporter returns the following `Organizations` metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_organization_info | Labeled Cloud Foundry Organization information with a constant `1` value | `environment`, `deployment`, `organization_id`, `organization_name`, `quota_name` |
| *metrics.namespace*_organization_non_basic_services_allowed | A Cloud Foundry Organization can provision instances of paid service plans? (`1` for `true`, `0` for `false`) | `environment`, `deployment`, `organization_id`, `organization_name` |
| *metrics.namespace*_organization_instance_memory_mb_limit | Maximum amount of memory (Mb) an application instance can have in a Cloud Foundry Organization | `environment`, `deployment`, `organization_id`, `organization_name` |
| *metrics.namespace*_organization_total_app_instances_quota | Total number of application instances that may be created in a Cloud Foundry Organization | `environment`, `deployment`, `organization_id`, `organization_name` |
| *metrics.namespace*_organization_total_app_tasks_quota | Total number of application tasks that may be created in a Cloud Foundry Organization | `environment`, `deployment`, `organization_id`, `organization_name` |
| *metrics.namespace*_organization_total_memory_mb_quota | Total amount of memory (Mb) a Cloud Foundry Organization can have | `environment`, `deployment`, `organization_id`, `organization_name` |
| *metrics.namespace*_organization_total_private_domains_quota | Total number of private domains that may be created in a Cloud Foundry Organization | `environment`, `deployment`, `organization_id`, `organization_name` |
| *metrics.namespace*_organization_total_reserved_route_ports_quota | Total number of routes that may be created with reserved ports in a Cloud Foundry Organization | `environment`, `deployment`, `organization_id`, `organization_name` |
| *metrics.namespace*_organization_total_routes_quota | Total number of routes that may be created in a Cloud Foundry Organization | `environment`, `deployment`, `organization_id`, `organization_name` |
| *metrics.namespace*_organization_total_service_keys_quota | Total number of service keys that may be created in a Cloud Foundry Organization | `environment`, `deployment`, `organization_id`, `organization_name` |
| *metrics.namespace*_organization_total_services_quota | Total number of service instances that may be created in a Cloud Foundry Organization | `environment`, `deployment`, `organization_id`, `organization_name` |
| *metrics.namespace*_organizations_scrapes_total | Total number of scrapes for Cloud Foundry Organizations | `environment`, `deployment` |
| *metrics.namespace*_organizations_scrape_errors_total | Total number of scrape errors of Cloud Foundry Organizations | `environment`, `deployment` |
| *metrics.namespace*_last_organizations_scrape_error | Whether the last scrape of Organizations metrics from Cloud Foundry resulted in an error (`1` for error, `0` for success) | `environment`, `deployment` |
| *metrics.namespace*_last_organizations_scrape_timestamp | Number of seconds since 1970 since last scrape of Organizations metrics from Cloud Foundry | `environment`, `deployment` |
| *metrics.namespace*_last_organizations_scrape_duration_seconds | Duration of the last scrape of Organizations metrics from Cloud Foundry | `environment`, `deployment` |

The exporter returns the following `Routes` metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_route_info | Labeled Cloud Foundry Route information with a constant `1` value | `environment`, `deployment`, `route_id`, `route_host`, `route_path`, `domain_id`, `space_id`, `service_instance_id` |
| *metrics.namespace*_routes_scrapes_total | Total number of scrapes for Cloud Foundry Routes | `environment`, `deployment` |
| *metrics.namespace*_routes_scrape_errors_total | Total number of scrape errors of Cloud Foundry Routes | `environment`, `deployment` |
| *metrics.namespace*_last_routes_scrape_error | Whether the last scrape of Routes metrics from Cloud Foundry resulted in an error (`1` for error, `0` for success) | `environment`, `deployment` |
| *metrics.namespace*_last_routes_scrape_timestamp | Number of seconds since 1970 since last scrape of Routes metrics from Cloud Foundry | `environment`, `deployment` |
| *metrics.namespace*_last_routes_scrape_duration_seconds | Duration of the last scrape of Routes metrics from Cloud Foundry | `environment`, `deployment` |


The exporter returns the following `Security Groups` metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_security_group_info | Labeled Cloud Foundry Security Group information with a constant `1` value | `environment`, `deployment`, `security_group_id`, `security_group_name` |
| *metrics.namespace*_security_groups_scrapes_total | Total number of scrapes for Cloud Foundry Security Groups | `environment`, `deployment` |
| *metrics.namespace*_security_groups_scrape_errors_total | Total number of scrape errors of Cloud Foundry Security Groups | `environment`, `deployment` |
| *metrics.namespace*_last_security_groups_scrape_error | Whether the last scrape of Security Groups metrics from Cloud Foundry resulted in an error (`1` for error, `0` for success) | `environment`, `deployment` |
| *metrics.namespace*_last_security_groups_scrape_timestamp | Number of seconds since 1970 since last scrape of Security Groups metrics from Cloud Foundry | `environment`, `deployment` |
| *metrics.namespace*_last_security_groups_scrape_duration_seconds | Duration of the last scrape of Security Groups metrics from Cloud Foundry | `environment`, `deployment` |

The exporter returns the following `Services` metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_service_info | Labeled Cloud Foundry Service information with a constant `1` value | `environment`, `deployment`, `service_id`, `service_label` |
| *metrics.namespace*_services_scrapes_total | Total number of scrapes for Cloud Foundry Services | `environment`, `deployment` |
| *metrics.namespace*_services_scrape_errors_total | Total number of scrape errors of Cloud Foundry Services | `environment`, `deployment` |
| *metrics.namespace*_last_services_scrape_error | Whether the last scrape of Services metrics from Cloud Foundry resulted in an error (`1` for error, `0` for success) | `environment`, `deployment` |
| *metrics.namespace*_last_services_scrape_timestamp | Number of seconds since 1970 since last scrape of Services metrics from Cloud Foundry | `environment`, `deployment` |
| *metrics.namespace*_last_services_scrape_duration_seconds | Duration of the last scrape of Services metrics from Cloud Foundry | `environment`, `deployment` |

The exporter returns the following `Service Instances` metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_service_instance_info | Labeled Cloud Foundry Service Instance information with a constant `1` value | `environment`, `deployment`, `service_instance_id`, `service_instance_name`, `service_plan_id`, `space_id`, `type` |
| *metrics.namespace*_service_instances_scrapes_total | Total number of scrapes for Cloud Foundry Service Instances | `environment`, `deployment` |
| *metrics.namespace*_service_instances_scrape_errors_total | Total number of scrape errors of Cloud Foundry Service Instances | `environment`, `deployment` |
| *metrics.namespace*_last_service_instances_scrape_error | Whether the last scrape of Service Instances metrics from Cloud Foundry resulted in an error (`1` for error, `0` for success) | `environment`, `deployment` |
| *metrics.namespace*_last_service_instances_scrape_timestamp | Number of seconds since 1970 since last scrape of Service Instances metrics from Cloud Foundry | `environment`, `deployment` |
| *metrics.namespace*_last_service_instances_scrape_duration_seconds | Duration of the last scrape of Service Instances metrics from Cloud Foundry | `environment`, `deployment` |

The exporter returns the following `Spaces` metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_space_info | Labeled Cloud Foundry Space information with a constant `1` value | `environment`, `deployment`, `space_id`, `space_name`, `organization_id`, `quota_name` |
| *metrics.namespace*_space_non_basic_services_allowed | A Cloud Foundry Space can provision instances of paid service plans? (`1` for `true`, `0` for `false`) | `environment`, `deployment`, `space_id`, `space_name`, `organization_id` |
| *metrics.namespace*_space_instance_memory_mb_limit | Maximum amount of memory (Mb) an application instance can have in a Cloud Foundry Space | `environment`, `deployment`, `space_id`, `space_name`, `organization_id` |
| *metrics.namespace*_space_total_app_instances_quota | Total number of application instances that may be created in a Cloud Foundry Space | `environment`, `deployment`, `space_id`, `space_name`, `organization_id` |
| *metrics.namespace*_space_total_app_tasks_quota | Total number of application tasks that may be created in a Cloud Foundry Space | `environment`, `deployment`, `space_id`, `space_name`, `organization_id` |
| *metrics.namespace*_space_total_memory_mb_quota | Total amount of memory (Mb) a Cloud Foundry Space can have | `environment`, `deployment`, `space_id`, `space_name`, `organization_id` |
| *metrics.namespace*_space_total_reserved_route_ports_quota | Total number of routes that may be created with reserved ports in a Cloud Foundry Space | `environment`, `deployment`, `space_id`, `space_name`, `organization_id` |
| *metrics.namespace*_space_total_routes_quota | Total number of routes that may be created in a Cloud Foundry Space | `environment`, `deployment`, `space_id`, `space_name`, `organization_id` |
| *metrics.namespace*_space_total_service_keys_quota | Total number of service keys that may be created in a Cloud Foundry Space | `environment`, `deployment`, `space_id`, `space_name`, `organization_id` |
| *metrics.namespace*_space_total_services_quota | Total number of service instances that may be created in a Cloud Foundry Space | `environment`, `deployment`, `space_id`, `space_name`, `organization_id` |
| *metrics.namespace*_spaces_scrapes_total | Total number of scrapes for Cloud Foundry Spaces | `environment`, `deployment` |
| *metrics.namespace*_spaces_scrape_errors_total | Total number of scrape errors of Cloud Foundry Spaces | `environment`, `deployment` |
| *metrics.namespace*_last_spaces_scrape_error | Whether the last scrape of Spaces metrics from Cloud Foundry resulted in an error (`1` for error, `0` for success) | `environment`, `deployment` |
| *metrics.namespace*_last_spaces_scrape_timestamp | Number of seconds since 1970 since last scrape of Spaces metrics from Cloud Foundry | `environment`, `deployment` |
| *metrics.namespace*_last_spaces_scrape_duration_seconds | Duration of the last scrape of Spaces metrics from Cloud Foundry | `environment`, `deployment` |

The exporter returns the following `Stacks` metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_stack_info | Labeled Cloud Foundry Stack information with a constant `1` value | `environment`, `deployment`, `stack_id`, `stack_name` |
| *metrics.namespace*_stacks_scrapes_total | Total number of scrapes for Cloud Foundry Stacks | `environment`, `deployment` |
| *metrics.namespace*_stacks_scrape_errors_total | Total number of scrape errors of Cloud Foundry Stacks | `environment`, `deployment` |
| *metrics.namespace*_last_stacks_scrape_error | Whether the last scrape of Stacks metrics from Cloud Foundry resulted in an error (`1` for error, `0` for success) | `environment`, `deployment` |
| *metrics.namespace*_last_stacks_scrape_timestamp | Number of seconds since 1970 since last scrape of Stacks metrics from Cloud Foundry | `environment`, `deployment` |
| *metrics.namespace*_last_stacks_scrape_duration_seconds | Duration of the last scrape of Stacks metrics from Cloud Foundry | `environment`, `deployment` |

## Acknowledgements

Thanks to [Michal Kuratczyk][mkuratczyk] who has also been working on a [cccf_exporter][cccf_exporter].

## Contributing

Refer to the [contributing guidelines][contributing].

## License

Apache License 2.0, see [LICENSE][license].

[binaries]: https://github.com/cloudfoundry-community/cf_exporter/releases
[cccf_exporter]: https://github.com/mkuratczyk/cfcc_exporter
[cloudfoundry]: https://www.cloudfoundry.org/
[contributing]: https://github.com/cloudfoundry-community/cf_exporter/blob/master/CONTRIBUTING.md
[faq]: https://github.com/cloudfoundry-community/cf_exporter/blob/master/FAQ.md
[golang]: https://golang.org/
[license]: https://github.com/cloudfoundry-community/cf_exporter/blob/master/LICENSE
[manifest]: https://github.com/cloudfoundry-community/cf_exporter/blob/master/manifest.yml
[mkuratczyk]: https://github.com/mkuratczyk
[prometheus]: https://prometheus.io/
[prometheus-boshrelease]: https://github.com/cloudfoundry-community/prometheus-boshrelease
