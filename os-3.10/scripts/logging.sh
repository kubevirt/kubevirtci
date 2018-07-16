#!/bin/bash

set -xe

MY_USER="---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fluentd
  namespace: logging
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: fluentd
  namespace: logging
rules:
- apiGroups:
  - ''
  resources:
  - pods
  - namespaces
  verbs:
  - get
  - list
  - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: fluentd
roleRef:
  kind: ClusterRole
  name: fluentd
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: fluentd
  namespace: logging
"

MY_LOGGING="---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: fluentd
  namespace: logging
  labels:
    k8s-app: fluentd-logging
    version: v1
    kubernetes.io/cluster-service: 'true'
spec:
  template:
    metadata:
      labels:
        k8s-app: fluentd-logging
        version: v1
        kubernetes.io/cluster-service: 'true'
    spec:
      serviceAccount: fluentd
      serviceAccountName: fluentd
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: fluentd
        image: pkotas/fluentd-daemonset:latest
        securityContext:
          privileged: true
        env:
          - name:  FLUENT_MASTER
            value: '192.168.66.2'
          - name:  FLUENT_PORT
            value: '24224'
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
        - name: configs
          mountPath: /fluentd/etc/
      terminationGracePeriodSeconds: 30
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      - name: configs
        configMap: 
          name: fluentd-daemonset"

/usr/bin/oc new-project logging
/usr/bin/oc project logging
echo "$MY_USER" | oc --config /etc/origin/master/admin.kubeconfig apply -f - 
/usr/bin/oc adm policy add-scc-to-user privileged system:serviceaccount:logging:fluentd
/usr/bin/oc adm policy add-cluster-role-to-user cluster-reader system:serviceaccount:logging:fluentd
/usr/bin/oc patch namespace kube-system -p '{"metadata": {"annotations": {"openshift.io/node-selector": ""}}}'
echo "$MY_LOGGING" | oc --config /etc/origin/master/admin.kubeconfig apply -f - 
