# FAQ

### What metrics does this exporter report?

The Cloud Foundry Prometheus Exporter gets the metrics from the [Cloud Foundry API][cf_api]. The metrics that are being [reported][cf_exporter_metrics] are enumerations of administrative information, and include:

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

### How can I enable only a particular collector?

The *filter.collectors* command flag allows you to filter what collectors will be enabled. Possible values are `Applications`, `Organizations`, `Services`, `Spaces` (or a combination of them).

### I have a question but I don't see it answered at this FAQ

We will be glad to address any questions not answered here. Please, just open a [new issue][issues].

[cf_api]: https://apidocs.cloudfoundry.org/
[cf_exporter_metrics]: https://github.com/cloudfoundry-community/cf_exporter#metrics
[firehose]: https://docs.cloudfoundry.org/loggregator/architecture.html#firehose
[firehose_exporter]: https://github.com/cloudfoundry-community/firehose_exporter
[issues]: https://github.com/cloudfoundry-community/cf_exporter/issues
