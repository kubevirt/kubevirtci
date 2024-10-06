package certs

import "net"

var ApiServerDnsNames = []string{"kubernetes", "kubernetes.default",
	"kubernetes.default.svc", "kubernetes.default.svc.cluster",
	"kubernetes.svc.cluster.local", "api-server.kubernetes.local"}

var ApiServerIPs = []net.IP{[]byte("127.0.0.1"), []byte("10.96.0.1"), []byte("192.168.66.110")}
