#!/bin/bash

set -xe

oc new-project logging
oc project logging
oc adm policy add-scc-to-user privileged system:serviceaccount:logging:fluentd
oc adm policy add-cluster-role-to-user cluster-reader system:serviceaccount:logging:fluentd
oc patch namespace kube-system -p '{"metadata": {"annotations": {"openshift.io/node-selector": ""}}}'
oc apply -f /tmp/logging.yaml
