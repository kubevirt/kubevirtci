FROM fedora:27

RUN dnf -y install glibc-static libtool zlib-static glib-devel libmnl-static libmnl-devel libnftnl-devel libnftnl make bzip2 bison flex && dnf clean all

ENV IPTABLES_VERSION 1.6.2

RUN cd / && curl -L -O ftp://ftp.netfilter.org/pub/iptables/iptables-${IPTABLES_VERSION}.tar.bz2 && ls && tar xf iptables-${IPTABLES_VERSION}.tar.bz2 && rm -rf iptables-1.6.2.tar.bz2

RUN cd iptables-${IPTABLES_VERSION} && sed -i '/if ENABLE_STATIC/a xtables_multi_LDFLAGS  = ${AM_LDFLAGS} -all-static' iptables/Makefile.am && autoreconf

RUN cd iptables-${IPTABLES_VERSION} && ./configure --disable-shared --enable-static --prefix=/build/usr \
   && make && make install
