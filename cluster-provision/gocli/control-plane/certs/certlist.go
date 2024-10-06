package certs

import (
	"crypto/x509"
	"net"
	"time"

	"k8s.io/client-go/util/cert"
)

type Cert struct {
	Name     string
	LongName string
	BaseName string
	CAName   string
	Config   cert.Config
	NotAfter *time.Time
}

type Certificates []*Cert

func GetAllCertificates() Certificates {
	return Certificates{
		ApiServerCert(),
		SchedulerCert(),
		ControllerMgrCert(),
		KubeletClientCert(),
		ServiceAccountsCert(),
		AdminCert(),
	}
}

func ApiServerCert() *Cert {
	return &Cert{
		Name:     "apiserver",
		LongName: "certificate for serving the Kubernetes API",
		BaseName: "kube-apiserver",
		CAName:   "ca",
		Config: cert.Config{
			CommonName: "kube-apiserver",
			AltNames: cert.AltNames{
				DNSNames: ApiServerDnsNames,
				IPs:      ApiServerIPs,
			},
			Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		},
	}
}

func SchedulerCert() *Cert {
	return &Cert{
		Name:     "kube-scheduler",
		LongName: "certificate for the kubernetes scheduler",
		BaseName: "kube-scheduler",
		CAName:   "ca",
		Config: cert.Config{
			CommonName:   "system:kube-scheduler",
			Organization: []string{"system:system:kube-scheduler"},
			AltNames: cert.AltNames{
				DNSNames: []string{"kube-scheduler"},
				IPs:      []net.IP{[]byte("127.0.0.1")},
			},
			Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		},
	}
}

func ControllerMgrCert() *Cert {
	return &Cert{
		Name:     "kube-controller-manager",
		LongName: "certificate for the kubernetes controller manager",
		BaseName: "kube-controller-manager",
		CAName:   "ca",
		Config: cert.Config{
			CommonName:   "system:kube-controller-manager",
			Organization: []string{"system:kube-controller-manager"},
			AltNames: cert.AltNames{
				DNSNames: []string{"kube-scheduler"},
				IPs:      []net.IP{[]byte("127.0.0.1")},
			},
			Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		},
	}
}

func ServiceAccountsCert() *Cert {
	return &Cert{
		Name:     "service-accounts",
		LongName: "certificate for kubernetes service account token signature",
		BaseName: "service-accounts",
		CAName:   "ca",
		Config: cert.Config{
			CommonName: "service-accounts",
		},
	}
}

func AdminCert() *Cert {
	return &Cert{
		Name:     "admin",
		LongName: "certificate for kubernetes adminstrator",
		BaseName: "admin",
		CAName:   "ca",
		Config: cert.Config{
			CommonName:   "admin",
			Organization: []string{"system:masters"},
		},
	}
}

func KubeletClientCert() *Cert {
	return &Cert{
		Name:     "apiserver-kubelet-client",
		LongName: "certificate for the API server to connect to kubelet",
		BaseName: "kube-apiserver-kubelet-client",
		CAName:   "ca",
		Config: cert.Config{
			CommonName:   "kube-apiserver-kubelet-client",
			Organization: []string{"system:masters"},
			Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		},
	}
}
