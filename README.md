# k8s clusters in qemu in docker

* `base` contains the base image with some scripts, qemu and dnsmasq
* `centos7` adds a vagrant centos7 box to the image
* `cli` contains a tool for provisioning, running and managing the containerized clusters
* `k8s-1.9.3` k8s-1.9.3 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.10.3` k8s-1.10.3 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.10.4` k8s-1.10.4 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.11.0` k8s-1.11.0 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-multus-1.10.4:`: k8s-1.10.4 cluster based on the centos7 image and uses multus CNI, provisioned with kubeadm
* `k8s-multus-1.11.1:`: k8s-1.11.1 cluster based on the centos7 image and uses multus CNI, provisioned with kubeadm
* `os-3.9` os-3.9 cluster based on the centos7 image, provisioned with openshift-ansible
* `os-3.9-crio` os-3.9 cluster with CRI-O support based on the centos7 image, provisioned with openshift-ansible
* `os-3.10.0` os-3.10.0 cluster based on the centos7 image, provisioned with openshift-ansible
* `os-3.10-crio` os-3.10 cluster with CRI-O support based on the centos7 image, provisioned with openshift-ansible
* `os-3.10-multus` os-3.10 cluster with multus cni support based on the centos7 image, provisioned with openshift-ansible

## Versions to use

* `kubevirtci/cli`: `sha256:1dd015dea4f12e6dcb8e31be3eeb677fed96f290ef4a4892a33c43d666053536`
* `kubevirtci/gocli`: `sha256:df958c060ca8d90701a1b592400b33852029979ad6d5c1d9b79683033704b690`
* `kubevirtci/base`: `sha256:034de1a154409d87498050ccc281d398ce1a0fed32efdbd66d2041a99a46b322`
* `kubevirtci/centos:1804_02`: `sha256:70653d952edfb8002ab8efe9581d01960ccf21bb965a9b4de4775c8fbceaab39`
* `kubevirtci/os-3.9.0:`: `sha256:234b3ae5c335c9fa32fa3bc01d5833f8f4d45420d82a8f8b12adc02687eb88b1`
* `kubevirtci/os-3.9.0-crio:`: `sha256:107d03dad4da6957e28774b121a45e177f31d7b4ad43c6eab7b24d467e59e213`
* `kubevirtci/os-3.10.0:`: `sha256:cdc9f998e19915b28b5c5be1ccc4c6fa2c8336435f38a37855f75b206977cbc2`
* `kubevirtci/os-3.10.0-crio:`: `sha256:36cc53fd72fbd247810b38618c0f3fa058c17bcb6744b0512193eba0728ebf05`
* `kubevirtci/os-3.10.0-multus`: `sha256:13cfb87ef4a060565936c89cd8f0687645f1e9888cdec952bdb622c8baab5fba`
* `kubevirtci/k8s-1.9.3:`: `sha256:f6ffb23261fb8aa15ed45b8d17e1299e284ea75e1d2814ee6b4ec24ecea6f24b`
* `kubevirtci/k8s-1.10.3:`: `sha256:d6290260e7e6b84419984f12719cf592ccbe327373b8df76aa0481f8ec01d357`
* `kubevirtci/k8s-1.10.4:`: `sha256:ee6846957b58e1f56b240d9ba6410f082e4787a4c4f1e0d60f6b907b76146b3e`
* `kubevirtci/k8s-1.11.0:`: `sha256:80e722a040fd6bcc5410239139a2cf0d33e1d52fe82d3cf9bdf5f5c3d09cfa7a`
* `kubevirtci/k8s-multus-1.10.4:`: `sha256:c66c2e1fdae0f92093b3e462e81c9856d423a918ba4dcdfcf401da284ea2bda1`
* `kubevirtci/k8s-multus-1.11.1:`: `sha256:ce7b961a449e20cf0628fba4deb90a8f399271659a0dd98eecdd249ec6c9d4b0`

## Using gocli

`gocli` is a tiny go binary which helps managing the containerized clusters. It
ca be used from a docker images, so no need to install it. You can for instance
use a bash alias:

```bash
alias gocli="docker run --net=host --privileged --rm -it -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/gocli:latest"
gocli help
```

## Quickstart Kubernetes

### Start the cluster

Start a k8s cluster which contains of one master and two nodes:

```bash
gocli run --random-ports --nodes 3 --background kubevirtci/k8s-1.10.3
```

### Connect to the cluster

Find out the connection details of the cluster:

```bash
$ gocli ports k8s
33396
$ gocli scp /etc/kubernetes/admin.conf - > ./kubeconfig
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

or to permamently edit kubeconfig:

```bash
$ gocli ports k8s
33396
$ gocli scp /etc/kubernetes/admin.conf - > ./kubeconfig
$ kubectl --kubeconfig=./kubeconfig config set-cluster kubernetes --server=https://127.0.0.1:33396
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

## Quickstart OpenShift

### Start the cluster

Start a k8s cluster which contains of one master and two nodes:

```bash
gocli run --random-ports --nodes 2 --reverse --ocp-port 8443 --background kubevirtci/os-3.9.0
```

Note the extra `--reverse` flag. Normally we start the master first and nodes
register. In the case of openshift it is different. We first need to start the
nodes, so that openshift-ansible can reach out to the nodes from master.

Furter we added `--ocp-port 8443`. That is only required if you want to access
the openshift-web-console. More about that further below.

### Connect to the cluster

Find out the connection details of the cluster:

```bash
$ gocli ports k8s
8443
$ gocli scp /etc/origin/master/admin.kubeconfig - > ./kubeconfig
$ oc --kubeconfig=./kubeconfig config set-cluster node01:8443 --server=127.0.0.1:8443
$ oc --kubeconfig=./kubeconfig config set-cluster node01:8443 --insecure-skip-tls-verify=true
$ oc --kubeconfig ./kubeconfig get nodes
```

### Accessing the webconsole

Make sure that `node01` resolves to `127.0.0.1` and that you added `--ocp-port
8443` when creatin the cluster. If you did that, you can simply access the
webconsole at `https://127.0.0.1:8443`. The login credentials are
`admin:admin`.

The two preconditions are necessary to make the authentication redirects work.

### Destroy the cluster

```bash
$ gocli rm
```
