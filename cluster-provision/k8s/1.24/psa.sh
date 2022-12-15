#!/bin/bash
set -xe

rm /etc/kubernetes/psa.yaml
cat > /etc/kubernetes/psa.yaml <<EOF
apiVersion: apiserver.config.k8s.io/v1
kind: AdmissionConfiguration
plugins:
- name: PodSecurity
  configuration:
    apiVersion: pod-security.admission.config.k8s.io/v1beta1
    kind: PodSecurityConfiguration
    defaults:
      enforce: "restricted"
      enforce-version: "latest"
      audit: "restricted"
      audit-version: "latest"
      warn: "restricted"
      warn-version: "latest"
    exemptions:
      usernames: []
      runtimeClasses: []
      # Hopefuly this will not be needed in future. Add your favorite namespace to be ignored and your operator not broken
      namespaces: ["kube-system", "default", "istio-operator" ,"istio-system", "nfs-csi", "monitoring", "rook-ceph", "cluster-network-addons", "sonobuoy"]
EOF