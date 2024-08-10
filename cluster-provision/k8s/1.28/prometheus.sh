#!/bin/bash
# set -xe

function usage()
{
    echo    "Script to run an experiment"
    echo    ""
    echo -e "\t-h --help"
    echo -e "\t-a --alertmanager deploy prometheus alertmanager (default false)"
    echo -e "\t-g --grafana deploy grafana with dashboards (default false)"
    echo    ""
}

ALERTMANAGER="false"
GRAFANA="false"
while [ "$1" != "" ]; do
    PARAM=$1; shift
    VALUE=$1; shift
    case $PARAM in
        -h | --help)
            usage
            exit 0
            ;;
        -a | --alertmanager)
            ALERTMANAGER=$VALUE
            ;;
        -g | --grafana)
            GRAFANA=$VALUE
            ;;
        *)
            echo "ERROR: unknown parameter \"$PARAM\""
            usage
            exit 1
            ;;
    esac
done

# Deploy Prometheus operator
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/0namespace-namespace.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-clusterRoleBinding.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-clusterRole.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-serviceAccount.yaml
### Prometheus operator CRDs
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-0prometheusCustomResourceDefinition.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-0podmonitorCustomResourceDefinition.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-0probeCustomResourceDefinition.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-0prometheusruleCustomResourceDefinition.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-0thanosrulerCustomResourceDefinition.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-0alertmanagerCustomResourceDefinition.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-0alertmanagerConfigCustomResourceDefinition.yaml
### Prometheus operator deployment
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-service.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus-operator/prometheus-operator-deployment.yaml

while [[ $(kubectl --kubeconfig /etc/kubernetes/admin.conf -n monitoring get pods -l app.kubernetes.io/name=prometheus-operator -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do 
  echo "Waiting for prometheus operator to be Ready, sleeping 20s and rechecking" && sleep 20;
done

# Deploy Prometheus
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus/prometheus-clusterRole.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus/prometheus-clusterRoleBinding.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus/prometheus-roleBindingConfig.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus/prometheus-roleBindingSpecificNamespaces.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus/prometheus-roleConfig.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus/prometheus-roleSpecificNamespaces.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus/prometheus-serviceAccount.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus/prometheus-podDisruptionBudget.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus/prometheus-service.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/prometheus/prometheus-prometheus.yaml

# Deploy Monitors
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/monitors/kubernetes-serviceMonitorApiserver.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/monitors/kubernetes-serviceMonitorCoreDNS.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/monitors/kubernetes-serviceMonitorKubeControllerManager.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/monitors/kubernetes-serviceMonitorKubeScheduler.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/monitors/kubernetes-serviceMonitorKubelet.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/monitors/prometheus-operator-serviceMonitor.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/monitors/prometheus-serviceMonitor.yaml

# Deploy kube-state-metrics
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/kube-state-metrics/kube-state-metrics-clusterRole.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/kube-state-metrics/kube-state-metrics-clusterRoleBinding.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/kube-state-metrics/kube-state-metrics-prometheusRule.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/kube-state-metrics/kube-state-metrics-serviceAccount.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/kube-state-metrics/kube-state-metrics-serviceMonitor.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/kube-state-metrics/kube-state-metrics-service.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/kube-state-metrics/kube-state-metrics-deployment.yaml

# Deploy node-exporter
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/node-exporter/node-exporter-clusterRole.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/node-exporter/node-exporter-clusterRoleBinding.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/node-exporter/node-exporter-prometheusRule.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/node-exporter/node-exporter-serviceAccount.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/node-exporter/node-exporter-serviceMonitor.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/node-exporter/node-exporter-daemonset.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/node-exporter/node-exporter-service.yaml

# Deploy alertmanager
if [[ ($ALERTMANAGER != "false") &&  ($ALERTMANAGER != "FALSE") ]]; then
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/alertmanager/alertmanager-secret.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/alertmanager/alertmanager-serviceAccount.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/alertmanager/alertmanager-serviceMonitor.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/alertmanager/alertmanager-podDisruptionBudget.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/alertmanager/alertmanager-service.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/alertmanager/alertmanager-alertmanager.yaml

    # Deploy alertmanager-rules
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/alertmanager-rules/alertmanager-prometheusRule.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/alertmanager-rules/kube-prometheus-prometheusRule.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/alertmanager-rules/prometheus-operator-prometheusRule.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/alertmanager-rules/prometheus-prometheusRule.yaml
fi

# Deploy grafana
if [[ ($GRAFANA != "false") && ($GRAFANA != "FALSE") ]]; then
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/grafana/grafana-dashboardDatasources.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/grafana/grafana-dashboardDefinitions.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/grafana/grafana-dashboardSources.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/grafana/grafana-deployment.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/grafana/grafana-service.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/grafana/grafana-serviceAccount.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/grafana/grafana-serviceMonitor.yaml
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/prometheus/grafana/grafana-config.yaml
fi

# Deploy nodeports
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/nodeports/monitoring.yaml
