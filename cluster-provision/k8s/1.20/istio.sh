#!/bin/bash
set -xe

source /var/lib/kubevirtci/shared_vars.sh

export PATH=/opt/istio-$ISTIO_VERSION/bin:$PATH

kubectl --kubeconfig /etc/kubernetes/admin.conf create ns istio-system
istioctl --kubeconfig /etc/kubernetes/admin.conf install --skip-confirmation \
  --set profile=demo \
  --set components.cni.enabled=true \
  --set values.cni.chained=true \
  --set values.cni.cniBinDir=/opt/cni/bin \
  --set values.cni.cniConfDir=/etc/cni/net.d \
  --set values.global.jwtPolicy=first-party-jwt
