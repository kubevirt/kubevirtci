#!/bin/bash
set -xe

source /var/lib/kubevirtci/shared_vars.sh

export PATH=/opt/istio-$ISTIO_VERSION/bin:$PATH

kubectl --kubeconfig /etc/kubernetes/admin.conf create ns istio-system
istioctl --kubeconfig /etc/kubernetes/admin.conf operator init

kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f - <<EOF
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  namespace: istio-system
  name: istio-operator
spec:
  profile: demo
  components:
    cni:
      enabled: true
      namespace: kube-system
  values:
    global:
      jwtPolicy: first-party-jwt
    cni:
      chained: false
      cniBinDir: /opt/cni/bin
      cniConfDir: /etc/cni/multus/net.d
      cniConfFileName: "istio-cni.conf"
      excludeNamespaces:
       - istio-system
       - kube-system
      logLevel: debug 
    sidecarInjectorWebhook:
      injectedAnnotations:
        "k8s.v1.cni.cncf.io/networks": istio-cni
EOF

retries=0
while [[ $retries -lt 20 ]]; do
  echo "waiting for istio-cni-node daemonset"
  sleep 5
  kubectl --kubeconfig /etc/kubernetes/admin.conf get daemonset istio-cni-node -n kube-system && break
  retries=$((retries + 1))
done

