FROM fluent/fluentd:v1.2.1-debian-onbuild
MAINTAINER Petr Kotas <petr.kotas@gmail.com>

RUN buildDeps="sudo make gcc g++ libc-dev ruby-dev" \
 && apt-get update \
 && apt-get install -y --no-install-recommends $buildDeps \
 && sudo gem install \
        fluent-plugin-secure-forward \
        fluent-plugin-record-reformer \
        fluent-plugin-kubernetes_metadata_filter \
        ffi \
        fluent-plugin-systemd \
        fluent-plugin-record-reformer \
 && sudo gem sources --clear-all \
 && SUDO_FORCE_REMOVE=yes \
    apt-get purge -y --auto-remove \
                  -o APT::AutoRemove::RecommendsImportant=false \
                  $buildDeps \
 && rm -rf /var/lib/apt/lists/* \
           /home/fluent/.gem/ruby/2.3.0/cache/*.gem


COPY kubernetes.conf /fluentd/etc/kubernetes.conf
COPY systemd.conf /fluentd/etc/systemd.conf
COPY entrypoint.sh  /bin/
RUN chmod +x /bin/entrypoint.sh

CMD exec fluentd -c /fluentd/etc/${FLUENTD_CONF} -p /fluentd/plugins $FLUENTD_OPT -v
