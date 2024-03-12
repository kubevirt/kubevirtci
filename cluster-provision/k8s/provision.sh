#!/bin/bash
# DO NOT RUN THIS SCRIPT STANDALONE, THIS GETS RUN AS PART OF THE PROVIDER SPECIFIC SCRIPTS. e.g: (cd cluster-provision/k8s/1.29; ../provision.sh)

set -ex

PHASES_DEFAULT="linux,k8s"
PHASES="${PHASES:-$PHASES_DEFAULT}"
CHECK_CLUSTER="${CHECK_CLUSTER:-false}"
export SLIM="${SLIM:-false}"
BYPASS_PMAN_CHANGE_CHECK=${BYPASS_PMAN_CHANGE_CHECK:-false}

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
provision_dir="$(basename "$(pwd)")"
base="$(cat base | tr -d '\n')"
export base

cd $DIR

make -C ../gocli cli

if [[ $BYPASS_PMAN_CHANGE_CHECK == false ]]; then
  json=$(cd ../.. && cluster-provision/gocli/build/cli provision-manager)
  result=$(echo $json | jq --arg v "$provision_dir" '.[$v]')
  if [[ $result == false ]]; then
    echo "INFO: skipping provision of $provision_dir because according provision-manager it hadn't changed"
    echo "INFO: use 'export BYPASS_PMAN_CHANGE_CHECK=true' to force provision"
    exit 0
  fi
fi

export KUBEVIRT_CGROUPV2="${CGROUPV2}"
if [[ $PHASES =~ linux.* ]]; then
  (cd ../${base} && ./build.sh)
fi

SLIM_MODE=""
if ${SLIM}; then
  SLIM_MODE="--slim"
fi

../gocli/build/cli provision ${provision_dir} --phases ${PHASES} ${SLIM_MODE}

if [[ $CHECK_CLUSTER != true ]] || [[ $PHASES == "linux" ]]; then
  echo "skipping cluster check when running linux only phase"
  exit 0
fi

./check-cluster-up.sh ${provision_dir}
