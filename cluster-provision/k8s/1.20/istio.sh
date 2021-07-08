#!/bin/bash
set -xe

source /var/lib/kubevirtci/shared_vars.sh

export PATH=$ISTIO_BIN_DIR:$PATH

kubectl --kubeconfig /etc/kubernetes/admin.conf create ns istio-system
istioctl --kubeconfig /etc/kubernetes/admin.conf operator init

istio_manifests_dir=/opt/istio
mkdir -p /opt/istio
cat <<EOF >$istio_manifests_dir/istio-operator.tpl.yaml
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
      chained: \$ISTIO_CNI_CHAINED
      cniBinDir: /opt/cni/bin
      cniConfDir: \$ISTIO_CNI_CONF_DIR
      excludeNamespaces:
       - istio-system
       - kube-system
      logLevel: debug
EOF

# generate istio-operator for usage with cnao enabled
ISTIO_CNI_CHAINED=false ISTIO_CNI_CONF_DIR=/etc/cni/multus/net.d envsubst < $istio_manifests_dir/istio-operator.tpl.yaml > $istio_manifests_dir/istio-operator-with-cnao.cr.yaml
cat <<EOF >>$istio_manifests_dir/istio-operator-with-cnao.yaml
      cniConfFileName: "istio-cni.conf"
    sidecarInjectorWebhook:
      injectedAnnotations:
        "k8s.v1.cni.cncf.io/networks": istio-cni
EOF

# generate istio-operator cr for usage without cnao
ISTIO_CNI_CHAINED=true ISTIO_CNI_CONF_DIR=/etc/cni/net.d envsubst < $istio_manifests_dir/istio-operator.tpl.yaml > $istio_manifests_dir/istio-operator.cr.yaml
