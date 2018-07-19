FROM fluent/fluentd:v1.2.1-onbuild
MAINTAINER Petr Kotas <petr.kotas@gmail.com>

# This image is here to distribute the config
# other option is to provide config via mount volume
# which will make gocli either have:
#  - one more option ie. `agregator-conf-dir`
#  - will have the config path hard coded thus resulting
#    gocli being dependent of particular folder structure

CMD exec fluentd -c /fluentd/etc/${FLUENTD_CONF} -p /fluentd/plugins $FLUENTD_OPT -v
