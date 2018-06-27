# Daemonset layer of agregated logging

Fluentd provided in this image is based on official fluentd daemonset image
provided by fluentd [here](https://github.com/fluent/fluentd-kubernetes-daemonset/blob/master/fluentd-daemonset-syslog.yaml).

However the original image is to general and it outputs to syslog as default.
It also does not log audit messages, which is one of the design goals
of the aggregated logging.

Custom image presented here drops the syslog extensions and gives and
oportunity to introduce further customizations.

The daemonset is configured to forward every log to aggregated endpoint.
