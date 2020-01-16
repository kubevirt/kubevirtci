#!/bin/sh

set -ex

TAG="$1"  #TODO: validate tag is vX.Y.Z
if [ -z "$TAG" ]; then
	TAG="devel"
fi

VERSIONDIR="internal/pkg/version"
VERSIONFILE="${VERSIONDIR}/version.go"

mkdir -p ${VERSIONDIR} && ./hack/build/genver.sh ${TAG} > ${VERSIONFILE}

export GO111MODULE=on
export GOPROXY=off
export GOFLAGS=-mod=vendor
go build -v .
