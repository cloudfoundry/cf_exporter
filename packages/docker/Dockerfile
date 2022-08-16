FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Ferran Rodenas <frodenas@gmail.com>

COPY cf_exporter /bin/cf_exporter

ENTRYPOINT ["/bin/cf_exporter"]
EXPOSE     9193