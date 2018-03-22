# k8s clusters in qemu in docker

* `base` contains the base image with some scripts, qemu and dnsmasq
* `centos7` adds a vagrant centos7 box to the image
* `cli` contains a tool for provisioning, running and managing the containerized clusters
* `k8s-1.9.3` k8s-1.9.3 cluster based on the centos7 image, provisioned with kubeadm
* `os-3.9` os-3.9 cluster based on the centos7 image, provisioned with openshift-ansible

## Versions to use

* `kubevirtci/cli`: `sha256:04ccd0e7da458fdb4947937df07ecce1431b01c89edeb0ed0e4a11c5036639a6`
* `kubevirtci/base`: `sha256:271136a97955e390e8eda02f3145f7bb89533e4023ed3a82c554f20e3b633549`
* `kubevirtci/centos:1608_01`: `sha256:94268aff21bb3b02f176b6ccbef0576d466ad31a540ca7269d6f99d31464081a`
* `kubevirtci/os-3.9:`: `sha256:b9ebf4517b82079bd0839d556530cdc6d9fab1f16e41b636696c953a03b890ce`
* `kubevirtci/k8s-1.9.3:`: `sha256:2f1600681800f70de293d2d35fa129bfd2c64e14ea01bab0284e4cafcc330662`
