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
* `kubevirtci/gocli`: `sha256:2ff1e9cddfa2cfdf268301a52d1a5ec252ace6908196609e001844e5458b746a`
* `kubevirtci/base`: `sha256:034de1a154409d87498050ccc281d398ce1a0fed32efdbd66d2041a99a46b322`
* `kubevirtci/centos:1804_02`: `sha256:70653d952edfb8002ab8efe9581d01960ccf21bb965a9b4de4775c8fbceaab39`
* **Deprecated**: `kubevirtci/os-3.9.0:`: `sha256:234b3ae5c335c9fa32fa3bc01d5833f8f4d45420d82a8f8b12adc02687eb88b1`
* **Deprecated**: `kubevirtci/os-3.9.0-crio:`: `sha256:107d03dad4da6957e28774b121a45e177f31d7b4ad43c6eab7b24d467e59e213`
* `kubevirtci/os-3.10.0:`: `sha256:b026dba96571a6732171c45b1f9fbdbbb3c3bbb1aa2118e99e12368c15ffb6f6`
* `kubevirtci/os-3.10.0-crio:`: `sha256:f79ae12ab7e0934c0786fcd08632c5b372163459b3e6c10aa85b329c1085e8e3`
* `kubevirtci/os-3.10.0-multus`: `sha256:7254ea84efcf8ae12ea0a97cbf5f2118803197ca9edfb422efe32dc3c1332fa0`
* **Deprecated**: `kubevirtci/k8s-1.9.3:`: `sha256:f6ffb23261fb8aa15ed45b8d17e1299e284ea75e1d2814ee6b4ec24ecea6f24b`
* **Deprecated**: `kubevirtci/k8s-1.10.3:`: `sha256:d6290260e7e6b84419984f12719cf592ccbe327373b8df76aa0481f8ec01d357`
* `kubevirtci/k8s-1.10.4:`: `sha256:c340a67190364b0e0c5864a8ce8edf38ccc35af6c4284a56118b2c38adf619cd`
* `kubevirtci/k8s-1.11.0:`: `sha256:39c1567b7caeae3f4845454bfeb2730a7d32abcac4a8dcedb62729c3248dd49c`
* `kubevirtci/k8s-multus-1.10.4:`: `sha256:a2ff7cef71441a1a7839d65e64bf4e2180b17b04bc1fa5203c2a8e98d77ac1d0`
* `kubevirtci/k8s-multus-1.11.1:`: `sha256:fe44e74ef12ae23e40d2a8a4e864c3a70bbb26923b5eeec5d365d69b13010ef8`

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
