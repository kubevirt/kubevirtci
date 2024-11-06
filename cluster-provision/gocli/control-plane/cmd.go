package controlplane

func buildEtcdCmdArgs() map[string]string {
	return map[string]string{
		"--advertise-client-urls":                       "https://0.0.0.0:2379",
		"--cert-file":                                   "/etc/kubernetes/pki/etcd/apiserver.crt",
		"--client-cert-auth":                            "true",
		"--data-dir":                                    "/var/lib/etcd",
		"--experimental-initial-corrupt-check":          "true",
		"--experimental-watch-progress-notify-interval": "5s",
		"--initial-advertise-peer-urls":                 "https://0.0.0.0:2380",
		"--initial-cluster":                             "node01=https://0.0.0.0:2380",
		"--key-file":                                    "/etc/kubernetes/pki/etcd/apiserver.pem",
		"--listen-client-urls":                          "https://0.0.0.0:2379",
		"--listen-metrics-urls":                         "http://0.0.0.0:2381",
		"--name":                                        "node01",
		"--snapshot-count":                              "10000",
		"--trusted-ca-file":                             "/etc/kubernetes/pki/etcd/ca.crt",
	}
}

func buildControllerMgrCmdArgs() map[string]string {
	return map[string]string{
		"--allocate-node-cidrs":              "true",
		"--authorization-kubeconfig":         "/etc/kubernetes/pki/kube-controller-manager/.kubeconfig",
		"--authentication-kubeconfig":        "/etc/kubernetes/pki/kube-controller-manager/.kubeconfig",
		"--bind-address":                     "127.0.0.1",
		"--cluster-cidr":                     "10.244.0.0/16,fd10:244::/112",
		"--cluster-name":                     "kubernetes",
		"--cluster-signing-cert-file":        "/etc/kubernetes/pki/ca.crt",
		"--cluster-signing-key-file":         "/etc/kubernetes/pki/key.pem",
		"--controllers":                      "*,csrapproving,csrsigning,bootstrapsigner,tokencleaner",
		"--kubeconfig":                       "/etc/kubernetes/pki/kube-controller-manager/.kubeconfig",
		"--node-cidr-mask-size-ipv6":         "116",
		"--leader-elect":                     "true",
		"-v":                                 "5",
		"--root-ca-file":                     "/etc/kubernetes/pki/ca.crt",
		"--service-account-private-key-file": "/etc/kubernetes/pki/service-accounts.pem",
		"--service-cluster-ip-range":         "10.96.0.0/12,fd10:96::/108",
		"--use-service-account-credentials":  "true",
	}
}

func buildApiServerCmdArgs() map[string]string {
	return map[string]string{
		"--advertise-address":                "192.168.66.110",
		"--allow-privileged":                 "true",
		"--audit-log-format":                 "json",
		"--audit-log-path":                   "/var/log/k8s-audit/k8s-audit.log",
		"--authorization-mode":               "Node,RBAC",
		"--client-ca-file":                   "/etc/kubernetes/pki/ca.crt",
		"--enable-admission-plugins":         "NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota",
		"--enable-bootstrap-token-auth":      "true",
		"--etcd-cafile":                      "/etc/kubernetes/pki/ca.crt",
		"--etcd-certfile":                    "/etc/kubernetes/pki/apiserver.crt",
		"--etcd-keyfile":                     "/etc/kubernetes/pki/apiserver.pem",
		"--etcd-servers":                     "https://127.0.0.1:2379",
		"--kubelet-preferred-address-types":  "InternalIP,ExternalIP,Hostname",
		"--secure-port":                      "6443",
		"--v":                                "3",
		"--service-account-issuer":           "https://kubernetes.default.svc.cluster.local",
		"--service-account-key-file":         "/etc/kubernetes/pki/service-accounts.pem",
		"--service-account-signing-key-file": "/etc/kubernetes/pki/service-accounts.pem",
		"--service-cluster-ip-range":         "10.96.0.0/24",
		"--tls-cert-file":                    "/etc/kubernetes/pki/apiserver.crt",
		"--tls-private-key-file":             "/etc/kubernetes/pki/apiserver.pem",
	}
}
