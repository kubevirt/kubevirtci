#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

PHASES_DEFAULT="linux,k8s"
PHASES="${PHASES:-$PHASES_DEFAULT}"
CHECK_CLUSTER="${CHECK_CLUSTER:-false}"

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
provision_dir="$(basename "$(pwd)")"
base="$(cat base | tr -d '\n')"
export base

cd $DIR

export KUBEVIRT_CGROUPV2="${CGROUPV2}"

if [[ $PHASES =~ linux.* ]]; then
  (cd ../${base} && ./build.sh)
fi

make -C ../gocli cli
../gocli/build/cli provision ${provision_dir} --phases ${PHASES}

if [[ $PHASES == $PHASES_DEFAULT ]] || [[ $CHECK_CLUSTER == true ]]; then
   if [[ $PHASES == "linux" ]]; then
     echo "skipping cluster check when running linux only phase"
     exit 0
   fi
  ./check-cluster-up.sh ${provision_dir}
fi
