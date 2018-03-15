# k8s clusters in qemu in docker

* `base` contains the base image with some scripts, qemu and dnsmasq
* `centos7` adds a vagrant centos7 box to the image
* `cli` contains a tool for provisioning, running and managing the containerized clusters
* `k8s-1.9.3` k8s-1.9.3 cluster based on the centos7 image, provisioned with kubeadm
* `os-3.9` os-3.9 cluster based on the centos7 image, provisioned with openshift-ansible

## Versions to use

* `kubevirtci/cli`: `sha256:c42004c9626e6a6e723a2107410694cb80864f3456725fdf945b1ca148ed6eaa`
* `kubevirtci/base`: `sha256:ab10913d74e7f157b9c18e4a270f707fc1ce9006a96448f97677b91ea36471a5`
* `kubevirtci/centos:1802_01`: `sha256:eeacdb20f0f5ec4e91756b99b9aa3e19287a6062bab5c3a41083cd245a44dc43`
* `kubevirtci/os-3.9:`: `sha256:a3c66710e0f4d55e81d5b2d32e89c074074cc14b216941818bde0d68cf4b0a12`
* `kubevirtci/k8s-1.9.3:`: `sha256:ead8cbdf16e205acfe66ec4b03e31974217e07808da1d9127409337d4959ace7`
