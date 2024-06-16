package prometheus

import (
	"embed"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

//go:embed manifests/*
var f embed.FS

type PrometheusOpt struct {
	grafanaEnabled      bool
	alertmanagerEnabled bool

	client k8s.K8sDynamicClient
}

func NewPrometheusOpt(c k8s.K8sDynamicClient, grafanaEnabled, alertmanagerEnabled bool) *PrometheusOpt {
	return &PrometheusOpt{
		grafanaEnabled:      grafanaEnabled,
		alertmanagerEnabled: alertmanagerEnabled,
		client:              c,
	}
}

func (o *PrometheusOpt) Exec() error {
	defaultManifests := []string{
		"manifests/prometheus-operator/0namespace-namespace.yaml",
		"manifests/prometheus-operator/prometheus-operator-clusterRoleBinding.yaml",
		"manifests/prometheus-operator/prometheus-operator-clusterRole.yaml",
		"manifests/prometheus-operator/prometheus-operator-serviceAccount.yaml",
		"manifests/prometheus-operator/prometheus-operator-0prometheusCustomResourceDefinition.yaml",
		"manifests/prometheus-operator/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml",
		"manifests/prometheus-operator/prometheus-operator-0podmonitorCustomResourceDefinition.yaml",
		"manifests/prometheus-operator/prometheus-operator-0probeCustomResourceDefinition.yaml",
		"manifests/prometheus-operator/prometheus-operator-0prometheusruleCustomResourceDefinition.yaml",
		"manifests/prometheus-operator/prometheus-operator-0thanosrulerCustomResourceDefinition.yaml",
		"manifests/prometheus-operator/prometheus-operator-0alertmanagerCustomResourceDefinition.yaml",
		"manifests/prometheus-operator/prometheus-operator-0alertmanagerConfigCustomResourceDefinition.yaml",
		"manifests/prometheus-operator/prometheus-operator-service.yaml",
		"manifests/prometheus-operator/prometheus-operator-deployment.yaml",
		"manifests/prometheus/prometheus-clusterRole.yaml",
		"manifests/prometheus/prometheus-clusterRoleBinding.yaml",
		"manifests/prometheus/prometheus-roleBindingConfig.yaml",
		"manifests/prometheus/prometheus-roleBindingSpecificNamespaces.yaml",
		"manifests/prometheus/prometheus-roleConfig.yaml",
		"manifests/prometheus/prometheus-roleSpecificNamespaces.yaml",
		"manifests/prometheus/prometheus-serviceAccount.yaml",
		"manifests/prometheus/prometheus-podDisruptionBudget.yaml",
		"manifests/prometheus/prometheus-service.yaml",
		"manifests/prometheus/prometheus-prometheus.yaml",
		"manifests/monitors/kubernetes-serviceMonitorApiserver.yaml",
		"manifests/monitors/kubernetes-serviceMonitorCoreDNS.yaml",
		"manifests/monitors/kubernetes-serviceMonitorKubeControllerManager.yaml",
		"manifests/monitors/kubernetes-serviceMonitorKubeScheduler.yaml",
		"manifests/monitors/kubernetes-serviceMonitorKubelet.yaml",
		"manifests/monitors/prometheus-operator-serviceMonitor.yaml",
		"manifests/monitors/prometheus-serviceMonitor.yaml",
		"manifests/kube-state-metrics/kube-state-metrics-clusterRole.yaml",
		"manifests/kube-state-metrics/kube-state-metrics-clusterRoleBinding.yaml",
		"manifests/kube-state-metrics/kube-state-metrics-prometheusRule.yaml",
		"manifests/kube-state-metrics/kube-state-metrics-serviceAccount.yaml",
		"manifests/kube-state-metrics/kube-state-metrics-serviceMonitor.yaml",
		"manifests/kube-state-metrics/kube-state-metrics-service.yaml",
		"manifests/kube-state-metrics/kube-state-metrics-deployment.yaml",
		"manifests/node-exporter/node-exporter-clusterRole.yaml",
		"manifests/node-exporter/node-exporter-clusterRoleBinding.yaml",
		"manifests/node-exporter/node-exporter-prometheusRule.yaml",
		"manifests/node-exporter/node-exporter-serviceAccount.yaml",
		"manifests/node-exporter/node-exporter-serviceMonitor.yaml",
		"manifests/node-exporter/node-exporter-daemonset.yaml",
		"manifests/node-exporter/node-exporter-service.yaml",
	}

	for _, manifest := range defaultManifests {
		err := o.client.Apply(f, manifest)
		if err != nil {
			return err
		}
	}

	if o.alertmanagerEnabled {
		alertmanagerManifests := []string{
			"manifests/alertmanager/alertmanager-secret.yaml",
			"manifests/alertmanager/alertmanager-serviceAccount.yaml",
			"manifests/alertmanager/alertmanager-serviceMonitor.yaml",
			"manifests/alertmanager/alertmanager-podDisruptionBudget.yaml",
			"manifests/alertmanager/alertmanager-service.yaml",
			"manifests/alertmanager/alertmanager-alertmanager.yaml",
			"manifests/alertmanager-rules/alertmanager-prometheusRule.yaml",
			"manifests/alertmanager-rules/kube-prometheus-prometheusRule.yaml",
			"manifests/alertmanager-rules/prometheus-operator-prometheusRule.yaml",
			"manifests/alertmanager-rules/prometheus-prometheusRule.yaml",
		}
		for _, manifest := range alertmanagerManifests {
			err := o.client.Apply(f, manifest)
			if err != nil {
				return err
			}
		}
	}

	if o.grafanaEnabled {
		grafanaManifests := []string{
			"manifests/grafana/grafana-dashboardDatasources.yaml",
			"manifests/grafana/grafana-dashboardDefinitions.yaml",
			"manifests/grafana/grafana-dashboardSources.yaml",
			"manifests/grafana/grafana-deployment.yaml",
			"manifests/grafana/grafana-service.yaml",
			"manifests/grafana/grafana-serviceAccount.yaml",
			"manifests/grafana/grafana-serviceMonitor.yaml",
		}

		for _, manifest := range grafanaManifests {
			err := o.client.Apply(f, manifest)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
