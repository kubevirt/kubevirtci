FROM fedora:27

RUN dnf -y install glibc-static libtool zlib-static glib-devel pcre-static libxml2-static xz-static libseccomp-static python2 glib2-devel glib2-static pixman-devel make bzip2  && dnf clean all

RUN cd / && curl -L -O http://xorg.freedesktop.org/archive/individual/lib/pixman-0.34.0.tar.bz2 && \
    tar xf pixman-0.34.0.tar.bz2 && \
    rm pixman-0.34.0.tar.bz2 && \
    cd pixman-0.34.0 && \
    ./configure --disable-ssse3 && \
    make && make install && \
    cd / && rm pixman* -rf

ENV QEMU_VERSION v2.10.2

RUN cd / && curl -L -O https://github.com/qemu/qemu/archive/${QEMU_VERSION}.tar.gz && tar xf ${QEMU_VERSION}.tar.gz && rm ${QEMU_VERSION}.tar.gz

RUN cd /qemu-2.10.2/ && ./configure --static --target-list=x86_64-softmmu --enable-kvm --disable-werror --prefix=/build/usr \
    && make install -j4
