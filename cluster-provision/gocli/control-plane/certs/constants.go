package certs

import "net"

var ApiServerDnsNames = []string{"kubernetes", "kubernetes.default",
	"kubernetes.default.svc", "kubernetes.default.svc.cluster",
	"kubernetes.svc.cluster.local", "api-server.kubernetes.local"}

var ApiServerIPs = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("10.96.0.1"), net.ParseIP("192.168.66.110")}
