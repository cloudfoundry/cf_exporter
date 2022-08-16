FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Ferran Rodenas <frodenas@gmail.com>

ADD cf_exporter /bin/cf_exporter

ENTRYPOINT ["/bin/cf_exporter"]
EXPOSE     9193
