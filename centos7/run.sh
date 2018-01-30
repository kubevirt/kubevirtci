#!/bin/bash

set -e

DNSMASQ_CID=$(docker run -d -p 5901:5901 -p 2222:22 -e NUM_NODES=3 --privileged qemu /dnsmasq.sh)
function finish() {
    docker stop ${DNSMASQ_CID}
    docker rm ${DNSMASQ_CID}
}
trap finish EXIT

test -t 1 && USE_TTY="-it"
docker run --rm ${USE_TTY} --privileged --net=container:${DNSMASQ_CID} qemu /bin/bash
