# FAQ

### What metrics does this exporter report?

The Cloud Foundry Prometheus Exporter gets information from the [Cloud Foundry API][cf_api]. The metrics that are being [reported][cf_exporter_metrics] are enumerations of administrative information, and include:

* Applications information:
  * Application information (id, name, space id and name, organization id and name)
  * Total number of applications
* Organizations information:
  * Organization information (id, name)
  * Total number of organizations
* Services information:
  * Service information (name, label)
  * Total number of services
* Spaces information:
  * Space information (id, name)
  * Total number of spaces

### How can I get detailed application metrics like CPU & Memory?

The goal of this exporter is just to provide administrative information about your Cloud Foundry environment. If you want to get detailed runtime application metrics, then you will need to use a different exporter, specifically, the [Cloud Foundry Firehose Prometheus Exporter][firehose_exporter], who will get `Container Metrics` from the [Cloud Foundry Firehose][firehose].

### Can I combine labels from a different exporter to get readable names?

Yes. You can combine this exporter with another exporter as far as there is a metric matching label.

For example, if you want to combine the `Container Metrics` from the [Cloud Foundry Firehose Prometheus Exporter][firehose_exporter] with this exporter you can run a query like:

```
firehose_container_metric_cpu_percentage
  * on(application_id)
  group_left(application_name, organization_name, space_name)
  cf_application_info
```

The *on* specifies the matching label, in this case, the *application_id*. The *group_left* specifies what labels (*application_name*, *organization_name*, *space_name*) from the right metric (*cf_application_info*) should be merged into the left metric (*firehose_container_metric_cpu_percentage*).

### How can I enable only a particular collector?

The *filter.collectors* command flag allows you to filter what collectors will be enabled. Possible values are `Applications`, `Organizations`, `Services`, `Spaces` (or a combination of them).

### Can I target multiple Cloud Foundry API endpoints with a single exporter instance?

No, this exporter only supports targetting a single [Cloud Foundry API][cf_api] endpoint. If you want to get metrics from several endpoints, you will need to use one exporter per endpoint.

### I have a question but I don't see it answered at this FAQ

We will be glad to address any questions not answered here. Please, just open a [new issue][issues].

[cf_api]: https://apidocs.cloudfoundry.org/
[cf_exporter_metrics]: https://github.com/cloudfoundry-community/cf_exporter#metrics
[firehose]: https://docs.cloudfoundry.org/loggregator/architecture.html#firehose
[firehose_exporter]: https://github.com/cloudfoundry-community/firehose_exporter
[issues]: https://github.com/cloudfoundry-community/cf_exporter/issues
