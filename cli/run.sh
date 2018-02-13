#!/bin/bash

PROVISION=false
RUN=false
NODES=0

while true; do
  case "$1" in
    -p | --provision ) PROVISION=true; shift ;;
    -n | --nodes ) NODES="$2"; shift 2 ;;
    -s | --scripts ) SCRIPTS="$2"; shift 2 ;;
    -t | --tag ) TAG="$2"; shift 2 ;;
    -b | --base ) BASE="$2"; shift 2 ;;
    -- ) shift; break ;;
    * ) break ;;
  esac
done

if [ -z "${BASE}" ]; then echo "Base image is not set. Use  '-b or --base' to set it."; exit 1; fi

if [ "$PROVISION" == "true" ] ; then
  if [ -z "${TAG}" ]; then echo "Resultin build tag not set. Use  '-t or --tag' to set it."; exit 1; fi
  if [ -z "${SCRIPTS}" ]; then echo "Provision script is not set. Use  '-s or --script' to set it."; exit 1; fi
fi

set -e

function finish() {
    set +e
    finish_dhcp 
    finish_vm
    set -e
}

trap finish EXIT

DNSMASQ_CID=$(docker run -d -p 5910:5901 -p 2201:2201 -e NUM_NODES=1 --privileged ${BASE} /bin/bash -c /dnsmasq.sh)
function finish_dhcp() {
    docker stop ${DNSMASQ_CID}
    docker rm ${DNSMASQ_CID}
}

if [ "$PROVISION" == "true" ] ; then
  VM_CID=$(docker run -d --privileged --net=container:${DNSMASQ_CID} ${BASE} /vm.sh --provision)
  function finish_vm() {
    docker stop ${VM_CID}
    docker rm ${VM_CID}
  }
  test -t 1 && USE_TTY="-it"
  docker cp ${SCRIPTS} ${VM_CID}:/scripts
  docker logs ${VM_CID}
  docker exec ${USE_TTY} ${VM_CID} /bin/bash -c "ssh.sh sudo /bin/bash < /scripts/provision.sh"
  docker exec ${USE_TTY} ${VM_CID} ssh.sh "sudo shutdown -h"
  docker wait ${VM_CID}
  docker commit --change "ENV PROVISIONED TRUE" ${VM_CID} ${TAG}
else
  NODE_NUM=01
  VM_CID=$(docker run -d --privileged -e NODE_NUM=${NODE_NUM} --net=container:${DNSMASQ_CID} ${BASE} /bin/bash -c /vm.sh)
  function finish_vm() {
    docker stop ${VM_CID}
    docker rm ${VM_CID}
  }
  docker exec ${USE_TTY} ${VM_CID} /bin/bash -c "ssh.sh sudo /bin/bash < /scripts/node${NODE_NUM}.sh"
  docker wait ${VM_CID}
fi
