# k8s clusters in QEMU in docker

* `base` contains the base image with some scripts, qemu and dnsmasq
* `centos7` adds a vagrant centos7 box to the image
* `cli` contains a tool for provisioning, running and managing the containerized clusters
* `k8s-1.9.3` k8s-1.9.3 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.10.3` k8s-1.10.3 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.10.4` k8s-1.10.4 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.10.11` k8s-1.10.11 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.11.0` k8s-1.11.0 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-1.13.3` k8s-1.13.3 cluster based on the centos7 image, provisioned with kubeadm
* `k8s-multus-1.10.4` k8s-1.10.4 cluster based on the centos7 image and uses multus CNI, provisioned with kubeadm
* `k8s-multus-1.11.1` k8s-1.11.1 cluster based on the centos7 image and uses multus CNI, provisioned with kubeadm
* `k8s-multus-1.12.2` k8s-1.12.2 cluster based on the centos7 image and uses multus CNI, provisioned with kubeadm
* `k8s-multus-1.13.3` k8s-1.13.3 cluster based on the centos7 image and uses multus CNI, provisioned with kubeadm
* `k8s-genie-1.11.1` k8s-1.11.1 cluster based on the centos7 image and uses genie CNI, provisioned with kubeadm
* `os-3.9` os-3.9 cluster based on the centos7 image, provisioned with openshift-ansible
* `os-3.9-crio` os-3.9 cluster with CRI-O support based on the centos7 image, provisioned with openshift-ansible
* `os-3.10.0` os-3.10.0 cluster based on the centos7 image, provisioned with openshift-ansible
* `os-3.10-crio` os-3.10 cluster with CRI-O support based on the centos7 image, provisioned with openshift-ansible
* `os-3.10-multus` os-3.10 cluster with multus cni support based on the centos7 image, provisioned with openshift-ansible
* `os-3.11.0` os-3.11.0 cluster based on the centos7 image, provisioned with openshift-ansible
* `os-3.11-crio` os-3.11 cluster with CRI-O support based on the centos7 image, provisioned
* `os-3.11-multus` os-3.11 cluster with multus cni support based on the centos7 image, provisioned with openshift-ansible

## Versions to use

* `kubevirtci/cli`: `sha256:1dd015dea4f12e6dcb8e31be3eeb677fed96f290ef4a4892a33c43d666053536`
* `kubevirtci/gocli`: `sha256:8677844b0c66aa02182c9d2c70b2b3bbe6f0f049e8617b630278622ca2a4228d`
* `kubevirtci/base`: `sha256:034de1a154409d87498050ccc281d398ce1a0fed32efdbd66d2041a99a46b322`
* `kubevirtci/centos:1804_02`: `sha256:70653d952edfb8002ab8efe9581d01960ccf21bb965a9b4de4775c8fbceaab39`
* **Deprecated**: `kubevirtci/os-3.9.0:`: `sha256:234b3ae5c335c9fa32fa3bc01d5833f8f4d45420d82a8f8b12adc02687eb88b1`
* **Deprecated**: `kubevirtci/os-3.9.0-crio:`: `sha256:107d03dad4da6957e28774b121a45e177f31d7b4ad43c6eab7b24d467e59e213`
* **Deprecated**: `kubevirtci/os-3.10.0:`: `sha256:cc418c0c837d8e6c9a31a063762d9e4c8bfc70a1fcca10823b11c6d8a7ae2394`
* **Deprecated**: `kubevirtci/os-3.10.0-crio:`: `sha256:56debd7bc2ce87dd616ebc30f06971e388b6983c0cda8646a7563e1dafadb69b`
* **Deprecated**: `kubevirtci/os-3.10.0-multus`: `sha256:875c973099141ab2013aaf51ec2b35b5326be943ef1437e4a87e041405e724ca`
* `kubevirtci/os-3.11.0-multus`: `sha256:3e962ef5042e6e03a25eea263d7474dd63e65e809d8d7b6ac3cf349e10dc13d9`
* `kubevirtci/os-3.11.0:`: `sha256:2d0a8f59dfebe181f550c4fbcd90d491a56a7d642d761c32a3c7732644325c0b`
* `kubevirtci/os-3.11.0-crio:`: `sha256:3f11a6f437fcdf2d70de4fcc31e0383656f994d0d05f9a83face114ea7254bc0`
* **Deprecated**: `kubevirtci/k8s-1.9.3:`: `sha256:f6ffb23261fb8aa15ed45b8d17e1299e284ea75e1d2814ee6b4ec24ecea6f24b`
* **Deprecated**: `kubevirtci/k8s-1.10.3:`: `sha256:d6290260e7e6b84419984f12719cf592ccbe327373b8df76aa0481f8ec01d357`
* **Deprecated**: `kubevirtci/k8s-1.10.4:`: `sha256:2ed70abfa8f6c30d990b76b816577040c0709258cbbd7c70f71a70d547f5544f`
* `kubevirtci/k8s-1.10.11:`: `sha256:f563a8ab4719e53c2372c4f41dfe55256677ec7afc442dfaebd494926005e3e5`
* `kubevirtci/k8s-1.11.0:`: `sha256:4d28fb90ac31fad9788fb0316e3bfd268d7ad322bd1b442075183eacacd1ace2`
* `kubevirtci/k8s-1.13.3:`: `sha256:9c2b78e11c25b3fd0b24b0ed684a112052dff03eee4ca4bdcc4f3168f9a14396`
* **Deprecated**: `kubevirtci/k8s-multus-1.10.4:`: `sha256:1d16d436347fcb9eba28cad08f6e074d4628e9a097a7325eb1ab87351e7f6d5c`
* **Deprecated**: `kubevirtci/k8s-multus-1.11.1:`: `sha256:3d35b19105344e270be168920e98287fbefcd5366fdf78681712d05725204559`
* **Deprecated**: `kubevirtci/k8s-multus-1.12.2:`: `sha256:d3ebf4a59ed379e59552ce8165596ade121d32e6b8782c192ff4f29863547c78`
* `kubevirtci/k8s-multus-1.13.3:`: `sha256:8418163abbc4acddefa26a63706446af9fd67aab20388a5e8d453cff0c0169f1`
* `kubevirtci/k8s-genie-1.11.1:`: `sha256:f1e2e729302bc0f2d4a176807f36c02b44c62605ad80d10249bb85682433b3a4`

# OKD clusters in the container with libvirt

* `okd-base` contains all needed packages to provision and run OKD cluster on top of the libvirt provider
* `okd-4.1.0-rc.0` okd-4.1.0-rc.0 cluster provisioned with OpenShift installer on top of the libvirt provider

## Versions to use

* `kubevirtci/okd-base`: `sha256:918d3c7f7c5ec94057715897f589c11b38e74c80927ee5af857e24817baeebaf`
* `kubevirtci/okd-4.1.0-rc.0`: `sha256:71afc34d9299aa11313b43a584e7e9d7e2f962279453b53d853a9d3bcb8b3255`

## Using gocli

`gocli` is a tiny go binary which helps managing the containerized clusters. It
ca be used from a docker images, so no need to install it. You can for instance
use a bash alias:

```bash
alias gocli="docker run --net=host --privileged --rm -it -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/gocli:latest"
gocli help
```

### How to provision OKD cluster

First you will need to create installer pull token file with the content:
```
pullSecret: <pull secret>
```

and after you should run `gocli` command:
```bash
gocli provision okd \
--prefix okd-4.1.0-rc.0 \
--dir-scripts <scripts_folder>/scripts \
--dir-hacks <hacks_folder>/hacks \
--master-memory 10240 \
--installer-pull-token-file <installer_pull_token_file> \
--installer-repo-tag release-4.1 \
--installer-release-image quay.io/openshift-release-dev/ocp-release:4.1.0-rc.0 \
kubevirtci/okd-base@sha256:918d3c7f7c5ec94057715897f589c11b38e74c80927ee5af857e24817baeebaf
```

***
NOTE: you can get the pull secret [here](https://developers.redhat.com/auth/realms/rhd/protocol/openid-connect/auth?client_id=uhc&redirect_uri=https%3A%2F%2Fcloud.openshift.com%2Fclusters%2Finstall%23pull-secret&state=109aa48e-1779-45d6-9bdc-c156b1e699b4&response_mode=fragment&response_type=code&scope=openid&nonce=b9fe0085-f2c9-4fd7-bd17-e8629f01bbaf).
***

***
NOTE: OpenShift cluster consumes a lot of resources, you should have at least 18Gb of the memory on the machine where do you run the container.
***

### How to run OKD cluster

You should run `gocli` command:
```bash
gocli run okd --prefix okd-4.1.0-rc.0 --background kubevirtci/okd-4.1.0-rc.0@sha256:71afc34d9299aa11313b43a584e7e9d7e2f962279453b53d853a9d3bcb8b3255
```

## Quickstart Kubernetes

### Start the cluster

Start a k8s cluster which contains of one master and two nodes:

```bash
gocli run --random-ports --nodes 3 --background kubevirtci/k8s-1.13.3
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
$ gocli scp /etc/kubernetes/admin.conf - > ./kubeconfig
$ kubectl --kubeconfig=./kubeconfig config set-cluster kubernetes --server=https://127.0.0.1:$(gocli ports k8s|tr -d '\r\n')
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
gocli run --random-ports --nodes 2 --memory 5120M --reverse --ocp-port 8443 --background kubevirtci/os-3.11.0
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
