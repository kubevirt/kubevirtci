package utils

const (
	// NFSGaneshaImage contains the reference to NFS docker image
	NFSGaneshaImage = "docker.io/janeczku/nfs-ganesha@sha256:17fe1813fd20d9fdfa497a26c8a2e39dd49748cd39dbb0559df7627d9bcf4c53"
	// CephImage contains the reference to CEPH docker image
	CephImage = "docker.io/ceph/daemon@sha256:939b053df0d0c92a3df24426f1ec5c31bc263560b152417a060e7caf41c0cc7e"
	// DockerRegistryImage contains the reference to docker registry docker image
	DockerRegistryImage = "docker.io/library/registry:2.7.1"
	// FluentdImage contains the reference to fluentd docker image
	FluentdImage = "docker.io/fluent/fluentd:v1.2-debian"
)
