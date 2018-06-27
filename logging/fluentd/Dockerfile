FROM fluent/fluentd:v1.2.1-onbuild
MAINTAINER Petr Kotas <petr.kotas@gmail.com>

# Needed for the forwaring to work properly
EXPOSE 24224/udp

CMD exec fluentd -c /fluentd/etc/${FLUENTD_CONF} -p /fluentd/plugins $FLUENTD_OPT -v
