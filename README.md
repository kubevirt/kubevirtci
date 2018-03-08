# k8s clusters in qemu in docker

* `base` contains the base image with some scripts, qemu and dnsmasq
* `centos7` adds a vagrant centos7 box to the image
* `cli` contains a tool for provisioning, running and managing the containerized clusters
* `k8s-1.9.3` k8s-1.9.3 cluster based on the centos7 image, provisioned with kubeadm
* `os-3.9` os-3.9 cluster based on the centos7 image, provisioned with openshift-ansible

## Versions to use

* `kubevirtci/cli`: `sha256:e640c6a23e75eeb98ac8230c0542c83a0a4a556bed72d99fa89b3fb68c3976d5`
* `kubevirtci/base`: `sha256:25f7dac76b2c88d8b7e0b1d6213d3406e77c7f230bfa1e66bd1cbb81a944eaaf`
* `kubevirtci/centos:1608_01`: `sha256:bd2bf287ce3b28a3624575b5dd31e375bbb213502693c4723d7a945e12dcf0f8`
* `kubevirtci/os-3.9:`: `sha256:67f864198299f658ae7199263d99a8e03afed7d02b252d3ccd2ba6ec434eae4f`
* `kubevirtci/k8s-1.9.3:`: `sha256:972483a8f2a1f3d1a3e4a921e316766ace87a1ec39e22be4d6bd8e29187ec570`
