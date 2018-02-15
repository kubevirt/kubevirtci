# k8s clusters in qemu in docker

* `base` contains the base image with some scripts, qemu and dnsmasq
* `centos7` adds a vagrant centos7 box to the image
* `cli` contains a tool for provisioning, running and managing the containerized clusters
* `kubeadm-1.9.3` k8s-1.9.3 cluster based on the centos7 image, provisioned with kubeadm
