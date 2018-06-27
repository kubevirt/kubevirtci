# Agregation layer of fluentd logging

The fluentd provided in this container servers as agregation endpoint.
It agregates logs from other fluend loggers placed on each cluster host.

The fluentd is setup in a way it stores every aggregated log inside common
directory in one file. Other fluentd loggers have to be set to forward.

The fluentd runs on fixed ip and port: `192.168.66.2:24224`
