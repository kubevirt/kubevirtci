#!/bin/bash

PROVISION=false
NODES=1

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
else
  if [ "$NODES" -lt "1" ]; then echo "The number of nodes must be greater or equal to 1."; exit 1 ; fi
fi

set -e

function finish() {
    set +e
    for id in ${CONTAINERS}; do
      docker stop ${id}
      docker rm ${id}
    done
    set -e
}

trap finish EXIT SIGINT SIGTERM SIGQUIT

DNSMASQ_CID=$(docker run -d -p 5910:5901 -p 2201:2201 -e NUM_NODES=${NODES} --privileged ${BASE} /bin/bash -c /dnsmasq.sh)
CONTAINERS=${DNSMASQ_CID}

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
  for i in $(seq 1 ${NODES}); do
    NODE_NUM="$(printf "%02d" ${i})"
    VM_CID=$(docker run -d --privileged -e NODE_NUM=${NODE_NUM} --net=container:${DNSMASQ_CID} ${BASE} /bin/bash -c /vm.sh)
    CONTAINERS="${CONTAINERS} ${VM_CID}"
    if docker exec ${USE_TTY} ${VM_CID} /bin/bash -c "test -f /scripts/node${NODE_NUM}.sh" ; then
      docker exec ${USE_TTY} ${VM_CID} /bin/bash -c "ssh.sh sudo /bin/bash < /scripts/node${NODE_NUM}.sh"
    elif docker exec ${USE_TTY} ${VM_CID} /bin/bash -c "test -f /scripts/nodes.sh" ; then
      docker exec ${USE_TTY} ${VM_CID} /bin/bash -c "ssh.sh sudo /bin/bash < /scripts/nodes.sh"
    fi
  done
  docker wait ${VM_CID} &
fi

wait
