package prometheus

import (
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/k8s/common"
)

type PrometheusOpt struct {
	grafanaEnabled      bool
	alertmanagerEnabled bool

	client *k8s.K8sDynamicClient
}

func NewPrometheusOpt(c *k8s.K8sDynamicClient, grafanaEnabled, alertmanagerEnabled bool) *PrometheusOpt {
	return &PrometheusOpt{
		grafanaEnabled:      grafanaEnabled,
		alertmanagerEnabled: alertmanagerEnabled,
		client:              c,
	}
}

func (o *PrometheusOpt) Exec() error {
	defaultManifests := []string{
		"/workdir/manifests/prometheus/prometheus-operator/0namespace-namespace.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-clusterRoleBinding.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-clusterRole.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-serviceAccount.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-0prometheusCustomResourceDefinition.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-0podmonitorCustomResourceDefinition.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-0probeCustomResourceDefinition.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-0prometheusruleCustomResourceDefinition.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-0thanosrulerCustomResourceDefinition.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-0alertmanagerCustomResourceDefinition.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-0alertmanagerConfigCustomResourceDefinition.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-service.yaml",
		"/workdir/manifests/prometheus/prometheus-operator/prometheus-operator-deployment.yaml",
		"/workdir/manifests/prometheus/prometheus/prometheus-clusterRole.yaml",
		"/workdir/manifests/prometheus/prometheus/prometheus-clusterRoleBinding.yaml",
		"/workdir/manifests/prometheus/prometheus/prometheus-roleBindingConfig.yaml",
		"/workdir/manifests/prometheus/prometheus/prometheus-roleBindingSpecificNamespaces.yaml",
		"/workdir/manifests/prometheus/prometheus/prometheus-roleConfig.yaml",
		"/workdir/manifests/prometheus/prometheus/prometheus-roleSpecificNamespaces.yaml",
		"/workdir/manifests/prometheus/prometheus/prometheus-serviceAccount.yaml",
		"/workdir/manifests/prometheus/prometheus/prometheus-podDisruptionBudget.yaml",
		"/workdir/manifests/prometheus/prometheus/prometheus-service.yaml",
		"/workdir/manifests/prometheus/prometheus/prometheus-prometheus.yaml",
		"/workdir/manifests/prometheus/monitors/kubernetes-serviceMonitorApiserver.yaml",
		"/workdir/manifests/prometheus/monitors/kubernetes-serviceMonitorCoreDNS.yaml",
		"/workdir/manifests/prometheus/monitors/kubernetes-serviceMonitorKubeControllerManager.yaml",
		"/workdir/manifests/prometheus/monitors/kubernetes-serviceMonitorKubeScheduler.yaml",
		"/workdir/manifests/prometheus/monitors/kubernetes-serviceMonitorKubelet.yaml",
		"/workdir/manifests/prometheus/monitors/prometheus-operator-serviceMonitor.yaml",
		"/workdir/manifests/prometheus/monitors/prometheus-serviceMonitor.yaml",
		"/workdir/manifests/prometheus/kube-state-metrics/kube-state-metrics-clusterRole.yaml",
		"/workdir/manifests/prometheus/kube-state-metrics/kube-state-metrics-clusterRoleBinding.yaml",
		"/workdir/manifests/prometheus/kube-state-metrics/kube-state-metrics-prometheusRule.yaml",
		"/workdir/manifests/prometheus/kube-state-metrics/kube-state-metrics-serviceAccount.yaml",
		"/workdir/manifests/prometheus/kube-state-metrics/kube-state-metrics-serviceMonitor.yaml",
		"/workdir/manifests/prometheus/kube-state-metrics/kube-state-metrics-service.yaml",
		"/workdir/manifests/prometheus/kube-state-metrics/kube-state-metrics-deployment.yaml",
		"/workdir/manifests/prometheus/node-exporter/node-exporter-clusterRole.yaml",
		"/workdir/manifests/prometheus/node-exporter/node-exporter-clusterRoleBinding.yaml",
		"/workdir/manifests/prometheus/node-exporter/node-exporter-prometheusRule.yaml",
		"/workdir/manifests/prometheus/node-exporter/node-exporter-serviceAccount.yaml",
		"/workdir/manifests/prometheus/node-exporter/node-exporter-serviceMonitor.yaml",
		"/workdir/manifests/prometheus/node-exporter/node-exporter-daemonset.yaml",
		"/workdir/manifests/prometheus/node-exporter/node-exporter-service.yaml",
	}

	for _, manifest := range defaultManifests {
		err := o.client.Apply(manifest)
		if err != nil {
			return err
		}
	}

	if o.alertmanagerEnabled {
		alertmanagerManifests := []string{
			"/tmp/prometheus/alertmanager/alertmanager-secret.yaml",
			"/tmp/prometheus/alertmanager/alertmanager-serviceAccount.yaml",
			"/tmp/prometheus/alertmanager/alertmanager-serviceMonitor.yaml",
			"/tmp/prometheus/alertmanager/alertmanager-podDisruptionBudget.yaml",
			"/tmp/prometheus/alertmanager/alertmanager-service.yaml",
			"/tmp/prometheus/alertmanager/alertmanager-alertmanager.yaml",
			"/tmp/prometheus/alertmanager-rules/alertmanager-prometheusRule.yaml",
			"/tmp/prometheus/alertmanager-rules/kube-prometheus-prometheusRule.yaml",
			"/tmp/prometheus/alertmanager-rules/prometheus-operator-prometheusRule.yaml",
			"/tmp/prometheus/alertmanager-rules/prometheus-prometheusRule.yaml",
		}
		for _, manifest := range alertmanagerManifests {
			err := o.client.Apply(manifest)
			if err != nil {
				return err
			}
		}
	}

	if o.grafanaEnabled {
		grafanaManifests := []string{
			"/tmp/prometheus/grafana/grafana-dashboardDatasources.yaml",
			"/tmp/prometheus/grafana/grafana-dashboardDefinitions.yaml",
			"/tmp/prometheus/grafana/grafana-dashboardSources.yaml",
			"/tmp/prometheus/grafana/grafana-deployment.yaml",
			"/tmp/prometheus/grafana/grafana-service.yaml",
			"/tmp/prometheus/grafana/grafana-serviceAccount.yaml",
			"/tmp/prometheus/grafana/grafana-serviceMonitor.yaml",
		}

		for _, manifest := range grafanaManifests {
			err := o.client.Apply(manifest)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
