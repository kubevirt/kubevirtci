# k8s clusters in qemu in docker

* `base` contains the base image with some scripts, qemu and dnsmasq
* `centos7` adds a vagrant centos7 box to the image
* `cli` contains a tool for provisioning, running and managing the containerized clusters
* `k8s-1.9.3` k8s-1.9.3 cluster based on the centos7 image, provisioned with kubeadm
* `os-3.9` os-3.9 cluster based on the centos7 image, provisioned with openshift-ansible

## Versions to use

* `kubevirtci/cli`: `sha256:b0023d1863338ef04fa0b8a8ee5956ae08616200d89ffd2e230668ea3deeaff4`
* `kubevirtci/gocli`: `sha256:367ab192305949b79294f92ceefa93a4647e53db0fcfd9ff9af6721af262fd55`
* `kubevirtci/base`: `sha256:67b84e2acefdcd7197989cbab1f2d1324eb87b5a77bd31d52d3000d13eee750c`
* `kubevirtci/centos:1802_01`: `sha256:31a48682e870c6eb9a60b26e49016f654238a1cb75127f2cca37b7eda29b05e5`
* `kubevirtci/os-3.9.0:`: `sha256:507263b67ddad581b086418d443c4fd1b2a30954e8fddff12254e0061a529410`
* `kubevirtci/k8s-1.9.3:`: `sha256:2ea1e772b13067617a7fc14f14105cfec5f97dd6e55db1827c3e877ba293fa8d`


## Using gocli

`gocli` is a tiny go binary which helps managing the containerized clusters. It
ca be used from a docker images, so no need to install it. You can for instance
use a bash alias:

```bash
alias gocli="docker run --privileged --rm -it -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/gocli:latest"
gocli help
```

## Quickstart k8s

### Start the cluster

Start a k8s cluster which contains of one master and two nodes:

```bash
gocli run --random-ports --nodes 3 --background kubevirtci/k8s-1.9.3
```

### Connect to the cluster

Find out the connection details of the cluster:

```bash
$ gocli ports k8s
33396
$ gocli ssh node01 -- sudo cat /etc/kubernetes/admin.conf | tail -n +3 > ./kubeconfig
$ kubectl --kubeconfig ./kubeconfig --insecure-skip-tls-verify --server https://localhost:33396 get pods -n kube-system
NAME                             READY     STATUS    RESTARTS   AGE
etcd-node01                      1/1       Running   0          14m
kube-apiserver-node01            1/1       Running   0          13m
kube-controller-manager-node01   1/1       Running   0          14m
kube-dns-6f4fd4bdf-mh6nb         3/3       Running   0          14m
kube-flannel-ds-4bk76            1/1       Running   0          14m
kube-flannel-ds-5zgmt            1/1       Running   1          14m
kube-flannel-ds-qbm2r            1/1       Running   1          14m
kube-proxy-gtvpb                 1/1       Running   0          14m
kube-proxy-knc6p                 1/1       Running   0          14m
kube-proxy-vx9t6                 1/1       Running   0          14m
kube-scheduler-node01            1/1       Running   0          13m
```

or to prepare the kubeconfig permanent:


```bash
$ gocli ports k8s
33396
$ gocli ssh node01 -- sudo cat /etc/kubernetes/admin.conf | tail -n +3 > ./kubeconfig
$ kubectl --kubeconfig=./kubeconfig config set-cluster kubernetes --server=https://127.0.0.1:$(_port k8s)
$ kubectl --kubeconfig=./kubeconfig config set-cluster kubernetes --insecure-skip-tls-verify=true
$ kubectl --kubeconfig ./kubeconfig get pods -n kube-system
NAME                             READY     STATUS    RESTARTS   AGE
etcd-node01                      1/1       Running   0          14m
kube-apiserver-node01            1/1       Running   0          13m
kube-controller-manager-node01   1/1       Running   0          14m
kube-dns-6f4fd4bdf-mh6nb         3/3       Running   0          14m
kube-flannel-ds-4bk76            1/1       Running   0          14m
kube-flannel-ds-5zgmt            1/1       Running   1          14m
kube-flannel-ds-qbm2r            1/1       Running   1          14m
kube-proxy-gtvpb                 1/1       Running   0          14m
kube-proxy-knc6p                 1/1       Running   0          14m
kube-proxy-vx9t6                 1/1       Running   0          14m
kube-scheduler-node01            1/1       Running   0          13m
```

### Destroy the cluster

```bash
$ gocli rm
```
