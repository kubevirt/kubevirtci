#!/bin/bash

set -ex

HUGEPAGES_2M=0
HUGEPAGES_1G=0

while true; do
  case "$1" in
    ----hugepages2M ) HUGEPAGES_2M="$2"; shift 2 ;;
    ----hugepages1G ) HUGEPAGES_1G="$2"; shift 2 ;;
    -- ) shift; break ;;
    * ) break ;;
  esac
done

if [ $HUGEPAGES_2M -ne 0 ]; then
  sudo sh -c "echo $(HUGEPAGES_2M) > /sys/devices/system/node/node0/hugepages/hugepages-2048kB/nr_hugepages"
fi

if [ $HUGEPAGES_1G -ne 0 ]; then
  sudo sh -c "echo $(HUGEPAGES_1G) > /sys/devices/system/node/node0/hugepages/hugepages-1048576kB/nr_hugepages"
fi

service kubelet restart
kubelet_rc=$?
if [[ $kubelet_rc -ne 0 ]]; then
    rm -rf /var/lib/kubelet/cpu_manager_state
    service kubelet restart
fi
