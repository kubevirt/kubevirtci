# Prometheus

> All the yamls here are based on the `release-0.8` of the repository [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus), a newer version may change significantly at any time.

The kube-prometheus repository collects Kubernetes manifests, [Grafana](http://grafana.com/) dashboards, and [Prometheus rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) combined with documentation and scripts to provide easy to operate end-to-end Kubernetes cluster monitoring with [Prometheus](https://prometheus.io/) using the Prometheus Operator.

Components included in this package:

* The [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
* Highly available [Prometheus](https://prometheus.io/)
* Highly available [Alertmanager](https://github.com/prometheus/alertmanager) (optional)
* [Prometheus node-exporter](https://github.com/prometheus/node_exporter)
* [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics)
* [Grafana](https://grafana.com/) (optional)

We provide a stack for cluster monitoring with pre-configured yamls to collect metrics from Kubernetes components, along with a set of grafana dashboards and prometheus alerting rules.


## Integrating KubeVirt with the prometheus-operator

Prometheus supports service discovery based `ServiceMonitors` that describe, discover and manage monitoring targets to be scraped by Prometheus. 

KubeVirt has a special service `kubevirt-prometheus-metrics`, that list all the KubeVirt system-components that expose Prometheus `metrics` at their endpoint. 
Therefore, prometheus-operator can make use of the `kubevirt-prometheus-metrics` service to automatically create the appropriate Prometheus config to monitor all KubeVirt system-components.

The `kubevirt-prometheus-metrics` service can then be discovered by the ServiceMonitor using label selectors.

KubeVirt’s virt-operator, by default, checks the existence of the MonitorNamespace and MonitorServiceAccount, and automatically creates a ServiceMonitor resource in the MonitorNamespace. Additionally, KubeVirt also appropriate role and rolebinding in KubeVirt’s namespace.

## Upgrading
All the files are based on the `release-0.8` of the repository [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus), the change applied was decreasing the Prometheus and Alertmanager replicas from 2 to 1.

Additionally, we included a new Grafana dashboard `kubevirt-control-plane.json` in [cluster-provision/k8s/1.21/manifests/prometheus/grafana/grafana-dashboardDefinitions.yaml](https://github.com/kubevirt/kubevirtci/cluster-provision/k8s/1.21/manifests/prometheus/grafana/grafana-dashboardDefinitions.yaml).